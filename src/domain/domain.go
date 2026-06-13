package domain

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

type UserRole string

const (
	RoleAdmin  UserRole = "admin"
	RoleVendor UserRole = "vendor"
)

// IsValid reports whether the role is one of the two allowed values.
func (r UserRole) IsValid() bool {
	return r == RoleAdmin || r == RoleVendor
}

var (
	ErrNotFound  = errors.New("not found")
	ErrConflict  = errors.New("conflict")
	ErrInvalid   = errors.New("invalid input")
	ErrForbidden = errors.New("forbidden")
)

type User struct {
	ID                        int64
	Username                  string
	Password                  string
	Email                     string
	ComissionPerSaleInPercent *int16
	Role                      UserRole
}

// CalcCommission computes the commission amount for an order total using
// the user's per-sale percentage. A user without a percentage (nil or 0)
// produces a zero commission.
//
// Mirrors CarShopApi/src/Services/OrderService.cs ProcessComission.
func (u User) CalcCommission(orderTotal decimal.Decimal) decimal.Decimal {
	if u.ComissionPerSaleInPercent == nil || *u.ComissionPerSaleInPercent <= 0 {
		return decimal.Zero
	}
	pct := decimal.NewFromInt(int64(*u.ComissionPerSaleInPercent)).Div(decimal.NewFromInt(100))
	return pct.Mul(orderTotal)
}

type Car struct {
	ID          int64
	New         bool
	Brand       string
	Model       string
	Year        int32
	Price       decimal.Decimal
	Color       string
	Km          int32
	Description string
	Images      []CarImage
}

type CarImage struct {
	ID    int64
	URL   string
	CarID *int64
}

type Order struct {
	ID           int64
	CustomerName string
	OrderDate    time.Time
	Total        decimal.Decimal
	VendorID     int64
	Items        []OrderItem
}

type OrderItem struct {
	ID       int64
	OrderID  int64
	CarID    int64
	Price    decimal.Decimal
	Discount decimal.Decimal
}

type VendorComission struct {
	ID                  int64
	VendorID            int64
	VendorName          string
	ComissionPercentage decimal.Decimal
	ComissionAmount     decimal.Decimal
	OrderID             int64
	OrderTotal          decimal.Decimal
}
