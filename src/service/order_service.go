package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"utfpr.edu.br/carshop-api/src/domain"
	"utfpr.edu.br/carshop-api/src/infra/sqlc"
)

type OrderService struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
}

func NewOrderService(pool *pgxpool.Pool) *OrderService {
	return &OrderService{pool: pool, q: sqlc.New(pool)}
}

// List returns every order or filters by vendorID when non-nil. Items are
// loaded for the whole page in one query (no N+1).
func (s *OrderService) List(ctx context.Context, vendorID *int64) ([]domain.Order, error) {
	var rows []sqlc.Order
	var err error
	if vendorID != nil {
		rows, err = s.q.ListOrdersByVendor(ctx, *vendorID)
	} else {
		rows, err = s.q.ListOrders(ctx)
	}
	if err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}
	if len(rows) == 0 {
		return []domain.Order{}, nil
	}

	ids := make([]int64, len(rows))
	for i, o := range rows {
		ids[i] = o.ID
	}
	items, err := s.q.ListOrderItemsByOrderIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("list order items: %w", err)
	}
	byOrder := make(map[int64][]domain.OrderItem, len(rows))
	for _, it := range items {
		byOrder[it.OrderID] = append(byOrder[it.OrderID], orderItemFromSQLC(it))
	}

	out := make([]domain.Order, len(rows))
	for i, o := range rows {
		out[i] = orderFromSQLC(o, byOrder[o.ID])
	}
	return out, nil
}

func (s *OrderService) Get(ctx context.Context, id int64) (domain.Order, error) {
	row, err := s.q.GetOrder(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Order{}, domain.ErrNotFound
		}
		return domain.Order{}, fmt.Errorf("get order: %w", err)
	}
	items, err := s.q.ListOrderItemsByOrderIDs(ctx, []int64{id})
	if err != nil {
		return domain.Order{}, fmt.Errorf("get order items: %w", err)
	}
	out := orderFromSQLC(row, orderItemsFromSQLC(items))
	return out, nil
}

func (s *OrderService) Create(ctx context.Context, in domain.Order) (domain.Order, error) {
	var created domain.Order
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		row, err := q.CreateOrder(ctx, sqlc.CreateOrderParams{
			CustomerName: in.CustomerName,
			OrderDate:    pgtype.Timestamptz{Time: in.OrderDate, Valid: true},
			Total:        in.Total,
			VendorID:     in.VendorID,
		})
		if err != nil {
			return fmt.Errorf("insert order: %w", err)
		}

		if err := s.insertItems(ctx, q, row.ID, in.Items); err != nil {
			return err
		}

		if err := s.upsertCommission(ctx, q, row); err != nil {
			return err
		}

		items, err := q.ListOrderItemsByOrderIDs(ctx, []int64{row.ID})
		if err != nil {
			return fmt.Errorf("reload order items: %w", err)
		}
		created = orderFromSQLC(row, orderItemsFromSQLC(items))
		return nil
	})
	if err != nil {
		return domain.Order{}, err
	}
	return created, nil
}

func (s *OrderService) Update(ctx context.Context, id int64, in domain.Order) (domain.Order, error) {
	var updated domain.Order
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		row, err := q.UpdateOrder(ctx, sqlc.UpdateOrderParams{
			ID:           id,
			CustomerName: in.CustomerName,
			OrderDate:    pgtype.Timestamptz{Time: in.OrderDate, Valid: true},
			Total:        in.Total,
			VendorID:     in.VendorID,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.ErrNotFound
			}
			return fmt.Errorf("update order: %w", err)
		}

		existingIDs, err := q.ListOrderItemIDsByOrderID(ctx, row.ID)
		if err != nil {
			return fmt.Errorf("list existing item ids: %w", err)
		}
		existing := make(map[int64]struct{}, len(existingIDs))
		for _, eid := range existingIDs {
			existing[eid] = struct{}{}
		}

		var (
			updIDs       []int64
			updCarIDs    []int64
			updPrices    []decimal.Decimal
			updDiscounts []decimal.Decimal
			toInsert     []domain.OrderItem
		)
		for _, it := range in.Items {
			if it.ID != 0 {
				if _, ok := existing[it.ID]; ok {
					updIDs = append(updIDs, it.ID)
					updCarIDs = append(updCarIDs, it.CarID)
					updPrices = append(updPrices, it.Price)
					updDiscounts = append(updDiscounts, it.Discount)
					continue
				}
			}
			toInsert = append(toInsert, it)
		}

		if len(updIDs) > 0 {
			if err := q.BulkUpdateOrderItems(ctx, sqlc.BulkUpdateOrderItemsParams{
				OrderID:   row.ID,
				Ids:       updIDs,
				CarIds:    updCarIDs,
				Prices:    updPrices,
				Discounts: updDiscounts,
			}); err != nil {
				return fmt.Errorf("bulk update order items: %w", err)
			}
		}
		if err := s.insertItems(ctx, q, row.ID, toInsert); err != nil {
			return err
		}

		if err := s.upsertCommission(ctx, q, row); err != nil {
			return err
		}

		items, err := q.ListOrderItemsByOrderIDs(ctx, []int64{row.ID})
		if err != nil {
			return fmt.Errorf("reload order items: %w", err)
		}
		updated = orderFromSQLC(row, orderItemsFromSQLC(items))
		return nil
	})
	if err != nil {
		return domain.Order{}, err
	}
	return updated, nil
}

func (s *OrderService) Delete(ctx context.Context, id int64) error {
	n, err := s.q.DeleteOrder(ctx, id)
	if err != nil {
		return fmt.Errorf("delete order: %w", err)
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *OrderService) insertItems(ctx context.Context, q *sqlc.Queries, orderID int64, items []domain.OrderItem) error {
	if len(items) == 0 {
		return nil
	}
	rows := make([]sqlc.BulkInsertOrderItemsParams, len(items))
	for i, it := range items {
		rows[i] = sqlc.BulkInsertOrderItemsParams{
			OrderID:  orderID,
			CarID:    it.CarID,
			Price:    it.Price,
			Discount: it.Discount,
		}
	}
	if _, err := q.BulkInsertOrderItems(ctx, rows); err != nil {
		return fmt.Errorf("bulk insert order items: %w", err)
	}
	return nil
}

// upsertCommission keeps at most one vendor_comissions row per order_id.
// Replaces the C# bug where every PUT inserted a fresh commission row
// (see CarShopApi/src/Services/OrderService.cs UpdateAsync -> ProcessComission).
func (s *OrderService) upsertCommission(ctx context.Context, q *sqlc.Queries, order sqlc.Order) error {
	vendor, err := q.GetUserByID(ctx, order.VendorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%w: vendor %d not found", domain.ErrInvalid, order.VendorID)
		}
		return fmt.Errorf("lookup vendor: %w", err)
	}
	dv := domain.User{ComissionPerSaleInPercent: vendor.ComissionPerSaleInPercent}
	amount := dv.CalcCommission(order.Total)

	var percentage decimal.Decimal
	if vendor.ComissionPerSaleInPercent != nil {
		percentage = decimal.NewFromInt(int64(*vendor.ComissionPerSaleInPercent))
	}

	if _, err := q.UpsertComissionByOrder(ctx, sqlc.UpsertComissionByOrderParams{
		VendorID:            vendor.ID,
		VendorName:          vendor.Username,
		ComissionPercentage: percentage,
		ComissionAmount:     amount,
		OrderID:             order.ID,
		OrderTotal:          order.Total,
	}); err != nil {
		return fmt.Errorf("upsert commission: %w", err)
	}
	return nil
}

func orderFromSQLC(o sqlc.Order, items []domain.OrderItem) domain.Order {
	var ts time.Time
	if o.OrderDate.Valid {
		ts = o.OrderDate.Time
	}
	return domain.Order{
		ID:           o.ID,
		CustomerName: o.CustomerName,
		OrderDate:    ts,
		Total:        o.Total,
		VendorID:     o.VendorID,
		Items:        items,
	}
}

func orderItemFromSQLC(i sqlc.OrderItem) domain.OrderItem {
	return domain.OrderItem{
		ID:       i.ID,
		OrderID:  i.OrderID,
		CarID:    i.CarID,
		Price:    i.Price,
		Discount: i.Discount,
	}
}

func orderItemsFromSQLC(rows []sqlc.OrderItem) []domain.OrderItem {
	out := make([]domain.OrderItem, len(rows))
	for i, r := range rows {
		out[i] = orderItemFromSQLC(r)
	}
	return out
}
