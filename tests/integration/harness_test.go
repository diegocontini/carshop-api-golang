package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	carshop "utfpr.edu.br/carshop-api"
	"utfpr.edu.br/carshop-api/src/controller"
	"utfpr.edu.br/carshop-api/src/infra/db"
	"utfpr.edu.br/carshop-api/src/service"
)

// Suite spins up a real Postgres container, runs goose migrations, and
// wires the full gin router so every test exercises the production stack
// from HTTP down to SQL. Tables are truncated between tests so each test
// starts on a clean slate with only the seeded admin user.
type Suite struct {
	suite.Suite

	container testcontainers.Container
	pool      *pgxpool.Pool
	router    *gin.Engine
	users     *service.UserService

	adminToken string
}

func TestSuite(t *testing.T) {
	gin.SetMode(gin.TestMode)
	suite.Run(t, new(Suite))
}

func (s *Suite) SetupSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	container, err := tcpostgres.Run(ctx, "postgres:17-alpine",
		tcpostgres.WithDatabase("test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	s.Require().NoError(err)
	s.container = container

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	s.Require().NoError(err)

	pool, err := db.Open(ctx, dsn)
	s.Require().NoError(err)
	s.pool = pool

	s.Require().NoError(db.Migrate(ctx, pool, carshop.MigrationsFS, "migrations"))

	jwt := service.NewJWTService("test-secret-that-is-at-least-32-chars-long!", "CarShopApi", "CarShopApiClients", 60)
	users := service.NewUserService(pool)
	cars := service.NewCarService(pool)
	orders := service.NewOrderService(pool)
	comm := service.NewComissionService(pool)
	s.users = users

	s.router = controller.BuildRouter(controller.Deps{
		JWT:       jwt,
		Users:     users,
		Cars:      cars,
		Orders:    orders,
		Comission: comm,
	})
}

func (s *Suite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
	if s.container != nil {
		_ = s.container.Terminate(context.Background())
	}
}

func (s *Suite) SetupTest() {
	ctx := context.Background()
	_, err := s.pool.Exec(ctx, `TRUNCATE users, cars, car_images, orders, order_items, vendor_comissions RESTART IDENTITY CASCADE`)
	s.Require().NoError(err)
	s.Require().NoError(s.users.SeedAdmin(ctx, "admin", "admin", "admin@localhost"))
	s.adminToken = s.login("admin", "admin")
}

func (s *Suite) do(method, path, token string, body any) (*http.Response, []byte) {
	var r *http.Request
	if body != nil {
		raw, err := json.Marshal(body)
		s.Require().NoError(err)
		r = httptest.NewRequest(method, path, bytes.NewReader(raw))
		r.Header.Set("Content-Type", "application/json")
	} else {
		r = httptest.NewRequest(method, path, nil)
	}
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)
	resp := w.Result()
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return resp, raw
}

func (s *Suite) login(username, password string) string {
	resp, body := s.do(http.MethodPost, "/api/v1/auth/token", "", map[string]string{
		"username": username,
		"password": password,
	})
	s.Require().Equal(http.StatusOK, resp.StatusCode, "login failed: %s", body)
	var tok struct {
		Token string `json:"token"`
	}
	s.Require().NoError(json.Unmarshal(body, &tok))
	s.Require().NotEmpty(tok.Token)
	return tok.Token
}

func (s *Suite) createVendor(username, password string, percent int16) int64 {
	resp, body := s.do(http.MethodPost, "/api/v1/user", s.adminToken, map[string]any{
		"username":                  username,
		"password":                  password,
		"email":                     username + "@local",
		"comissionPerSaleInPercent": percent,
		"role":                      "Vendor",
	})
	s.Require().Equal(http.StatusCreated, resp.StatusCode, "createVendor failed: %s", body)
	var u struct {
		ID int64 `json:"id"`
	}
	s.Require().NoError(json.Unmarshal(body, &u))
	return u.ID
}
