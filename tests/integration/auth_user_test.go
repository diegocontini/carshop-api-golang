package integration

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (s *Suite) TestLoginAsSeededAdmin() {
	resp, _ := s.do(http.MethodPost, "/api/v1/auth/token", "", map[string]string{
		"username": "admin", "password": "admin",
	})
	s.Equal(http.StatusOK, resp.StatusCode)
}

func (s *Suite) TestLoginWrongPasswordIs401() {
	resp, _ := s.do(http.MethodPost, "/api/v1/auth/token", "", map[string]string{
		"username": "admin", "password": "nope",
	})
	s.Equal(http.StatusUnauthorized, resp.StatusCode)
}

func (s *Suite) TestLoginMissingFieldsIs400() {
	resp, _ := s.do(http.MethodPost, "/api/v1/auth/token", "", map[string]string{
		"username": "admin",
	})
	s.Equal(http.StatusBadRequest, resp.StatusCode)
}

func (s *Suite) TestUserEndpointsRequireAuth() {
	resp, _ := s.do(http.MethodGet, "/api/v1/user", "", nil)
	s.Equal(http.StatusUnauthorized, resp.StatusCode)
}

func (s *Suite) TestVendorCannotListUsers() {
	s.createVendor("vendor1", "vpw", 5)
	vt := s.login("vendor1", "vpw")
	resp, _ := s.do(http.MethodGet, "/api/v1/user", vt, nil)
	s.Equal(http.StatusForbidden, resp.StatusCode)
}

func (s *Suite) TestCreateUserOmitsPasswordInResponse() {
	resp, body := s.do(http.MethodPost, "/api/v1/user", s.adminToken, map[string]any{
		"username": "ada",
		"password": "secret-123",
		"email":    "ada@local",
		"role":     "Vendor",
	})
	s.Equal(http.StatusCreated, resp.StatusCode)
	s.False(strings.Contains(string(body), "password"), "response leaked password field: %s", body)
	s.False(strings.Contains(string(body), "secret-123"), "response leaked plaintext password: %s", body)
}

func (s *Suite) TestDuplicateUsernameIs409() {
	first, _ := s.do(http.MethodPost, "/api/v1/user", s.adminToken, map[string]any{
		"username": "dup", "password": "pw", "email": "dup@local", "role": "Vendor",
	})
	s.Require().Equal(http.StatusCreated, first.StatusCode)
	second, _ := s.do(http.MethodPost, "/api/v1/user", s.adminToken, map[string]any{
		"username": "dup", "password": "pw", "email": "dup@local", "role": "Vendor",
	})
	s.Equal(http.StatusConflict, second.StatusCode)
}

func (s *Suite) TestListUsersDoesNotIncludePasswordHash() {
	resp, body := s.do(http.MethodGet, "/api/v1/user", s.adminToken, nil)
	s.Equal(http.StatusOK, resp.StatusCode)
	var users []map[string]any
	s.Require().NoError(json.Unmarshal(body, &users))
	s.Require().NotEmpty(users)
	for _, u := range users {
		_, present := u["password"]
		s.False(present, "user payload contained password field: %v", u)
	}
}
