package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (s *Suite) TestVendorCannotCreateCar() {
	s.createVendor("vc", "pw", 0)
	vt := s.login("vc", "pw")
	resp, _ := s.do(http.MethodPost, "/api/v1/car", vt, map[string]any{
		"new":   true,
		"brand": "Honda", "model": "Civic", "year": 2024,
		"price": 100000, "color": "black", "km": 0, "description": "",
	})
	s.Equal(http.StatusForbidden, resp.StatusCode)
}

func (s *Suite) TestCreateCarWithImages() {
	resp, body := s.do(http.MethodPost, "/api/v1/car", s.adminToken, map[string]any{
		"new":   true,
		"brand": "Honda", "model": "Civic", "year": 2024,
		"price": 99999.99, "color": "black", "km": 0, "description": "fresh",
		"images": []map[string]any{
			{"url": "https://img/1"},
			{"url": "https://img/2"},
		},
	})
	s.Require().Equal(http.StatusCreated, resp.StatusCode, string(body))

	var car struct {
		ID     int64 `json:"id"`
		Images []struct {
			ID    int64  `json:"id"`
			URL   string `json:"url"`
			CarID int64  `json:"carId"`
		} `json:"images"`
	}
	s.Require().NoError(json.Unmarshal(body, &car))
	s.NotZero(car.ID)
	s.Require().Len(car.Images, 2)
	for _, im := range car.Images {
		s.NotZero(im.ID)
		s.Equal(car.ID, im.CarID)
	}

	get, body := s.do(http.MethodGet, fmt.Sprintf("/api/v1/car/%d", car.ID), s.adminToken, nil)
	s.Equal(http.StatusOK, get.StatusCode, string(body))
}

func (s *Suite) TestUpdateCarUpsertsImages() {
	resp, body := s.do(http.MethodPost, "/api/v1/car", s.adminToken, map[string]any{
		"new":   true,
		"brand": "Toyota", "model": "Corolla", "year": 2023,
		"price": 50000, "color": "silver", "km": 0, "description": "",
		"images": []map[string]any{{"url": "https://old/1"}},
	})
	s.Require().Equal(http.StatusCreated, resp.StatusCode)
	var created struct {
		ID     int64 `json:"id"`
		Images []struct {
			ID  int64  `json:"id"`
			URL string `json:"url"`
		} `json:"images"`
	}
	s.Require().NoError(json.Unmarshal(body, &created))
	s.Require().Len(created.Images, 1)
	existingID := created.Images[0].ID

	updateBody := map[string]any{
		"new":   false,
		"brand": "Toyota", "model": "Corolla", "year": 2023,
		"price": 49000, "color": "silver", "km": 100, "description": "demo",
		"images": []map[string]any{
			{"id": existingID, "url": "https://updated/1"},
			{"url": "https://new/2"},
		},
	}
	put, body := s.do(http.MethodPut, fmt.Sprintf("/api/v1/car/%d", created.ID), s.adminToken, updateBody)
	s.Require().Equal(http.StatusOK, put.StatusCode, string(body))

	var updated struct {
		Images []struct {
			ID  int64  `json:"id"`
			URL string `json:"url"`
		} `json:"images"`
	}
	s.Require().NoError(json.Unmarshal(body, &updated))
	s.Require().Len(updated.Images, 2)

	urls := map[string]bool{}
	for _, im := range updated.Images {
		urls[im.URL] = true
	}
	s.True(urls["https://updated/1"], "updated URL missing")
	s.True(urls["https://new/2"], "newly inserted URL missing")
}
