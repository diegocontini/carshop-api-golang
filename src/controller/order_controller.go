package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"utfpr.edu.br/carshop-api/src/controller/dto"
	"utfpr.edu.br/carshop-api/src/service"
)

type OrderController struct {
	svc *service.OrderService
}

func NewOrderController(svc *service.OrderService) *OrderController {
	return &OrderController{svc: svc}
}

// Register installs CRUD routes. adminOnly gates DELETE per the C# contract;
// the other routes are open to Admin+Vendor by the rg-level middleware.
func (c *OrderController) Register(rg *gin.RouterGroup, adminOnly gin.HandlerFunc) {
	rg.GET("", c.list)
	rg.GET("/:id", c.getByID)
	rg.POST("", c.create)
	rg.PUT("/:id", c.update)
	rg.DELETE("/:id", adminOnly, c.delete)
}

func (c *OrderController) list(ctx *gin.Context) {
	var vendorID *int64
	if v := ctx.Query("vendorId"); v != "" {
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid vendorId"})
			return
		}
		vendorID = &parsed
	}
	orders, err := c.svc.List(ctx.Request.Context(), vendorID)
	if err != nil {
		writeError(ctx, err)
		return
	}
	resp := make([]dto.OrderResponse, len(orders))
	for i, o := range orders {
		resp[i] = dto.OrderToResponse(o)
	}
	ctx.JSON(http.StatusOK, resp)
}

func (c *OrderController) getByID(ctx *gin.Context) {
	id, ok := parseID(ctx)
	if !ok {
		return
	}
	o, err := c.svc.Get(ctx.Request.Context(), id)
	if err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, dto.OrderToResponse(o))
}

func (c *OrderController) create(ctx *gin.Context) {
	var req dto.OrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	created, err := c.svc.Create(ctx.Request.Context(), req.ToDomain())
	if err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, dto.OrderToResponse(created))
}

func (c *OrderController) update(ctx *gin.Context) {
	id, ok := parseID(ctx)
	if !ok {
		return
	}
	var req dto.OrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	updated, err := c.svc.Update(ctx.Request.Context(), id, req.ToDomain())
	if err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, dto.OrderToResponse(updated))
}

func (c *OrderController) delete(ctx *gin.Context) {
	id, ok := parseID(ctx)
	if !ok {
		return
	}
	if err := c.svc.Delete(ctx.Request.Context(), id); err != nil {
		writeError(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}
