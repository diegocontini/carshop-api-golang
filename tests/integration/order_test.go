package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// TestOrderUpdateDoesNotDuplicateCommission is the regression test for the
// C# bug where every PUT on an order inserted a fresh vendor_comissions
// row. After two updates we expect exactly one commission row whose
// amount tracks the latest total.
func (s *Suite) TestOrderUpdateDoesNotDuplicateCommission() {
	vendorID := s.createVendor("vendor-orders", "vpw", 3) // 3%

	carResp, carBody := s.do(http.MethodPost, "/api/v1/car", s.adminToken, map[string]any{
		"new":   true,
		"brand": "Ford", "model": "Focus", "year": 2024,
		"price": 80000, "color": "white", "km": 0, "description": "",
	})
	s.Require().Equal(http.StatusCreated, carResp.StatusCode, string(carBody))
	var car struct{ ID int64 `json:"id"` }
	s.Require().NoError(json.Unmarshal(carBody, &car))

	createBody := map[string]any{
		"customerName": "Buyer",
		"orderDate":    time.Now().UTC().Format(time.RFC3339),
		"total":        100000,
		"vendorId":     vendorID,
		"items": []map[string]any{
			{"carId": car.ID, "price": 100000, "discount": 0},
		},
	}
	orderResp, orderBody := s.do(http.MethodPost, "/api/v1/order", s.adminToken, createBody)
	s.Require().Equal(http.StatusCreated, orderResp.StatusCode, string(orderBody))
	var order struct{ ID int64 `json:"id"` }
	s.Require().NoError(json.Unmarshal(orderBody, &order))

	// First commission state: 3% of 100000 = 3000
	commResp, commBody := s.do(http.MethodGet, fmt.Sprintf("/api/v1/comission?vendorId=%d", vendorID), s.adminToken, nil)
	s.Require().Equal(http.StatusOK, commResp.StatusCode)
	var comms []map[string]any
	s.Require().NoError(json.Unmarshal(commBody, &comms))
	s.Require().Len(comms, 1, "expected 1 commission after create, got %d", len(comms))
	s.Equal("3000", asString(comms[0]["comissionAmount"]))

	// Update the order total. C# would have inserted ANOTHER commission row here.
	updateBody := map[string]any{
		"customerName": "Buyer",
		"orderDate":    time.Now().UTC().Format(time.RFC3339),
		"total":        200000,
		"vendorId":     vendorID,
		"items": []map[string]any{
			{"carId": car.ID, "price": 200000, "discount": 0},
		},
	}
	put, putBody := s.do(http.MethodPut, fmt.Sprintf("/api/v1/order/%d", order.ID), s.adminToken, updateBody)
	s.Require().Equal(http.StatusOK, put.StatusCode, string(putBody))

	commResp, commBody = s.do(http.MethodGet, fmt.Sprintf("/api/v1/comission?vendorId=%d", vendorID), s.adminToken, nil)
	s.Require().Equal(http.StatusOK, commResp.StatusCode)
	s.Require().NoError(json.Unmarshal(commBody, &comms))
	s.Require().Len(comms, 1, "expected exactly 1 commission after update (was the dup-row bug), got %d", len(comms))
	s.Equal("6000", asString(comms[0]["comissionAmount"]), "commission should reflect new total: 3%% of 200000")
}

func (s *Suite) TestListOrdersByVendor() {
	v1 := s.createVendor("va", "pw", 2)
	v2 := s.createVendor("vb", "pw", 4)

	car := s.makeCar(60000)

	s.makeOrder(v1, car, 60000)
	s.makeOrder(v1, car, 60000)
	s.makeOrder(v2, car, 60000)

	resp, body := s.do(http.MethodGet, fmt.Sprintf("/api/v1/order?vendorId=%d", v1), s.adminToken, nil)
	s.Require().Equal(http.StatusOK, resp.StatusCode)
	var orders []map[string]any
	s.Require().NoError(json.Unmarshal(body, &orders))
	s.Len(orders, 2, "expected 2 orders for v1")
}

func (s *Suite) makeCar(price int) int64 {
	resp, body := s.do(http.MethodPost, "/api/v1/car", s.adminToken, map[string]any{
		"new":   true,
		"brand": "X", "model": "Y", "year": 2024,
		"price": price, "color": "white", "km": 0, "description": "",
	})
	s.Require().Equal(http.StatusCreated, resp.StatusCode, string(body))
	var c struct{ ID int64 `json:"id"` }
	s.Require().NoError(json.Unmarshal(body, &c))
	return c.ID
}

func (s *Suite) makeOrder(vendorID, carID int64, total int) int64 {
	resp, body := s.do(http.MethodPost, "/api/v1/order", s.adminToken, map[string]any{
		"customerName": "Buyer",
		"orderDate":    time.Now().UTC().Format(time.RFC3339),
		"total":        total,
		"vendorId":     vendorID,
		"items": []map[string]any{
			{"carId": carID, "price": total, "discount": 0},
		},
	})
	s.Require().Equal(http.StatusCreated, resp.StatusCode, string(body))
	var o struct{ ID int64 `json:"id"` }
	s.Require().NoError(json.Unmarshal(body, &o))
	return o.ID
}

func asString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case json.Number:
		return x.String()
	default:
		raw, _ := json.Marshal(v)
		s := string(raw)
		if len(s) > 1 && s[0] == '"' && s[len(s)-1] == '"' {
			s = s[1 : len(s)-1]
		}
		return s
	}
}
