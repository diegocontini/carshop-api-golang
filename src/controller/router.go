package controller

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"utfpr.edu.br/carshop-api/src/controller/middleware"
	"utfpr.edu.br/carshop-api/src/domain"
	"utfpr.edu.br/carshop-api/src/service"
)

// Deps bundles every controller dependency BuildRouter needs.
type Deps struct {
	JWT       *service.JWTService
	Users     *service.UserService
	Cars      *service.CarService
	Orders    *service.OrderService
	Comission *service.ComissionService
}

// BuildRouter wires every endpoint under the same gin engine, applying the
// per-route auth middleware described in CarShopApi/src/Docs/routes.md.
func BuildRouter(d Deps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
		})
	})
	r.GET("/", func(c *gin.Context) { c.Redirect(http.StatusFound, "/docs/routes.md") })

	auth := middleware.RequireAuth(d.JWT)
	adminOrVendor := middleware.RequireRoles(domain.RoleAdmin, domain.RoleVendor)
	adminOnly := middleware.RequireRoles(domain.RoleAdmin)

	v1 := r.Group("/api/v1")

	NewAuthController(d.JWT, d.Users).Register(v1.Group("/auth"))

	users := v1.Group("/user", auth, adminOnly)
	NewUserController(d.Users).Register(users)

	cars := v1.Group("/car", auth, adminOrVendor)
	NewCarController(d.Cars).Register(cars, adminOnly)

	orders := v1.Group("/order", auth, adminOrVendor)
	NewOrderController(d.Orders).Register(orders, adminOnly)

	comm := v1.Group("/comission", auth, adminOrVendor)
	NewComissionController(d.Comission).Register(comm, adminOnly)

	return r
}
