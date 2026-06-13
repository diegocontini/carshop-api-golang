package dto

import (
	"github.com/shopspring/decimal"

	"utfpr.edu.br/carshop-api/src/domain"
)

type ComissionRequest struct {
	ID                  *int64          `json:"id"`
	VendorID            int64           `json:"vendorId"            binding:"required"`
	VendorName          string          `json:"vendorName"          binding:"required"`
	ComissionPercentage decimal.Decimal `json:"comissionPercentage"`
	ComissionAmount     decimal.Decimal `json:"comissionAmount"`
	OrderID             int64           `json:"orderId"             binding:"required"`
	OrderTotal          decimal.Decimal `json:"orderTotal"`
}

type ComissionResponse struct {
	ID                  int64           `json:"id"`
	VendorID            int64           `json:"vendorId"`
	VendorName          string          `json:"vendorName"`
	ComissionPercentage decimal.Decimal `json:"comissionPercentage"`
	ComissionAmount     decimal.Decimal `json:"comissionAmount"`
	OrderID             int64           `json:"orderId"`
	OrderTotal          decimal.Decimal `json:"orderTotal"`
}

func (r ComissionRequest) ToDomain() domain.VendorComission {
	return domain.VendorComission{
		VendorID:            r.VendorID,
		VendorName:          r.VendorName,
		ComissionPercentage: r.ComissionPercentage,
		ComissionAmount:     r.ComissionAmount,
		OrderID:             r.OrderID,
		OrderTotal:          r.OrderTotal,
	}
}

func ComissionToResponse(c domain.VendorComission) ComissionResponse {
	return ComissionResponse{
		ID:                  c.ID,
		VendorID:            c.VendorID,
		VendorName:          c.VendorName,
		ComissionPercentage: c.ComissionPercentage,
		ComissionAmount:     c.ComissionAmount,
		OrderID:             c.OrderID,
		OrderTotal:          c.OrderTotal,
	}
}
