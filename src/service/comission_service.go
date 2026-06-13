package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"utfpr.edu.br/carshop-api/src/domain"
	"utfpr.edu.br/carshop-api/src/infra/sqlc"
)

type ComissionService struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
}

func NewComissionService(pool *pgxpool.Pool) *ComissionService {
	return &ComissionService{pool: pool, q: sqlc.New(pool)}
}

func (s *ComissionService) List(ctx context.Context, vendorID *int64) ([]domain.VendorComission, error) {
	var rows []sqlc.VendorComission
	var err error
	if vendorID != nil {
		rows, err = s.q.ListComissionsByVendor(ctx, *vendorID)
	} else {
		rows, err = s.q.ListComissions(ctx)
	}
	if err != nil {
		return nil, fmt.Errorf("list comissions: %w", err)
	}
	out := make([]domain.VendorComission, len(rows))
	for i, r := range rows {
		out[i] = comissionFromSQLC(r)
	}
	return out, nil
}

func (s *ComissionService) Get(ctx context.Context, id int64) (domain.VendorComission, error) {
	row, err := s.q.GetComission(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.VendorComission{}, domain.ErrNotFound
		}
		return domain.VendorComission{}, fmt.Errorf("get comission: %w", err)
	}
	return comissionFromSQLC(row), nil
}

func (s *ComissionService) Create(ctx context.Context, in domain.VendorComission) (domain.VendorComission, error) {
	row, err := s.q.CreateComission(ctx, sqlc.CreateComissionParams{
		VendorID:            in.VendorID,
		VendorName:          in.VendorName,
		ComissionPercentage: in.ComissionPercentage,
		ComissionAmount:     in.ComissionAmount,
		OrderID:             in.OrderID,
		OrderTotal:          in.OrderTotal,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return domain.VendorComission{}, fmt.Errorf("%w: commission for this order already exists", domain.ErrConflict)
		}
		return domain.VendorComission{}, fmt.Errorf("create comission: %w", err)
	}
	return comissionFromSQLC(row), nil
}

func (s *ComissionService) Update(ctx context.Context, id int64, in domain.VendorComission) (domain.VendorComission, error) {
	row, err := s.q.UpdateComission(ctx, sqlc.UpdateComissionParams{
		ID:                  id,
		VendorID:            in.VendorID,
		VendorName:          in.VendorName,
		ComissionPercentage: in.ComissionPercentage,
		ComissionAmount:     in.ComissionAmount,
		OrderID:             in.OrderID,
		OrderTotal:          in.OrderTotal,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.VendorComission{}, domain.ErrNotFound
		}
		return domain.VendorComission{}, fmt.Errorf("update comission: %w", err)
	}
	return comissionFromSQLC(row), nil
}

func (s *ComissionService) Delete(ctx context.Context, id int64) error {
	n, err := s.q.DeleteComission(ctx, id)
	if err != nil {
		return fmt.Errorf("delete comission: %w", err)
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func comissionFromSQLC(c sqlc.VendorComission) domain.VendorComission {
	return domain.VendorComission{
		ID:                  c.ID,
		VendorID:            c.VendorID,
		VendorName:          c.VendorName,
		ComissionPercentage: c.ComissionPercentage,
		ComissionAmount:     c.ComissionAmount,
		OrderID:             c.OrderID,
		OrderTotal:          c.OrderTotal,
	}
}
