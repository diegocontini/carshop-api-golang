package dto

import (
	"time"

	"github.com/shopspring/decimal"

	"utfpr.edu.br/carshop-api/src/domain"
)

type OrderRequest struct {
	ID           *int64             `json:"id"`
	CustomerName string             `json:"customerName" binding:"required"`
	OrderDate    time.Time          `json:"orderDate"`
	Total        decimal.Decimal    `json:"total"`
	VendorID     int64              `json:"vendorId"     binding:"required"`
	Items        []OrderItemRequest `json:"items"`
}

type OrderItemRequest struct {
	ID       *int64          `json:"id"`
	CarID    int64           `json:"carId"    binding:"required"`
	Price    decimal.Decimal `json:"price"`
	Discount decimal.Decimal `json:"discount"`
}

type OrderResponse struct {
	ID           int64               `json:"id"`
	CustomerName string              `json:"customerName"`
	OrderDate    time.Time           `json:"orderDate"`
	Total        decimal.Decimal     `json:"total"`
	VendorID     int64               `json:"vendorId"`
	Items        []OrderItemResponse `json:"items"`
}

type OrderItemResponse struct {
	ID       int64           `json:"id"`
	OrderID  int64           `json:"orderId"`
	CarID    int64           `json:"carId"`
	Price    decimal.Decimal `json:"price"`
	Discount decimal.Decimal `json:"discount"`
}

func (r OrderRequest) ToDomain() domain.Order {
	items := make([]domain.OrderItem, len(r.Items))
	for i, it := range r.Items {
		var id int64
		if it.ID != nil {
			id = *it.ID
		}
		items[i] = domain.OrderItem{
			ID:       id,
			CarID:    it.CarID,
			Price:    it.Price,
			Discount: it.Discount,
		}
	}
	return domain.Order{
		CustomerName: r.CustomerName,
		OrderDate:    r.OrderDate,
		Total:        r.Total,
		VendorID:     r.VendorID,
		Items:        items,
	}
}

func OrderToResponse(o domain.Order) OrderResponse {
	items := make([]OrderItemResponse, len(o.Items))
	for i, it := range o.Items {
		items[i] = OrderItemResponse{
			ID:       it.ID,
			OrderID:  it.OrderID,
			CarID:    it.CarID,
			Price:    it.Price,
			Discount: it.Discount,
		}
	}
	return OrderResponse{
		ID:           o.ID,
		CustomerName: o.CustomerName,
		OrderDate:    o.OrderDate,
		Total:        o.Total,
		VendorID:     o.VendorID,
		Items:        items,
	}
}
