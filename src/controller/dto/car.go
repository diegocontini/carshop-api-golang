package dto

import (
	"github.com/shopspring/decimal"

	"utfpr.edu.br/carshop-api/src/domain"
)

type CarRequest struct {
	ID          *int64               `json:"id"`
	New         bool                 `json:"new"`
	Brand       string               `json:"brand"       binding:"required"`
	Model       string               `json:"model"       binding:"required"`
	Year        int32                `json:"year"        binding:"required"`
	Price       decimal.Decimal      `json:"price"`
	Color       string               `json:"color"       binding:"required"`
	Km          int32                `json:"km"`
	Description string               `json:"description"`
	Images      []CarImageRequest    `json:"images"`
}

type CarImageRequest struct {
	ID  *int64 `json:"id"`
	URL string `json:"url" binding:"required"`
}

type CarResponse struct {
	ID          int64              `json:"id"`
	New         bool               `json:"new"`
	Brand       string             `json:"brand"`
	Model       string             `json:"model"`
	Year        int32              `json:"year"`
	Price       decimal.Decimal    `json:"price"`
	Color       string             `json:"color"`
	Km          int32              `json:"km"`
	Description string             `json:"description"`
	Images      []CarImageResponse `json:"images"`
}

type CarImageResponse struct {
	ID    int64  `json:"id"`
	URL   string `json:"url"`
	CarID *int64 `json:"carId"`
}

// ToDomain converts the request to a domain.Car. Image ID is 0 when the
// client omitted it - the service interprets 0 as "insert new" and any
// non-zero ID as "update if it belongs to this car, else insert new"
// (matches the C# semantics).
func (r CarRequest) ToDomain() domain.Car {
	imgs := make([]domain.CarImage, len(r.Images))
	for i, im := range r.Images {
		var id int64
		if im.ID != nil {
			id = *im.ID
		}
		imgs[i] = domain.CarImage{ID: id, URL: im.URL}
	}
	return domain.Car{
		New:         r.New,
		Brand:       r.Brand,
		Model:       r.Model,
		Year:        r.Year,
		Price:       r.Price,
		Color:       r.Color,
		Km:          r.Km,
		Description: r.Description,
		Images:      imgs,
	}
}

func CarToResponse(c domain.Car) CarResponse {
	imgs := make([]CarImageResponse, len(c.Images))
	for i, im := range c.Images {
		imgs[i] = CarImageResponse{ID: im.ID, URL: im.URL, CarID: im.CarID}
	}
	return CarResponse{
		ID:          c.ID,
		New:         c.New,
		Brand:       c.Brand,
		Model:       c.Model,
		Year:        c.Year,
		Price:       c.Price,
		Color:       c.Color,
		Km:          c.Km,
		Description: c.Description,
		Images:      imgs,
	}
}

