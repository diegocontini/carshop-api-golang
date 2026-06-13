package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"utfpr.edu.br/carshop-api/src/controller/dto"
	"utfpr.edu.br/carshop-api/src/service"
)

type CarController struct {
	svc *service.CarService
}

func NewCarController(svc *service.CarService) *CarController {
	return &CarController{svc: svc}
}

// Register installs the CRUD routes on rg. The caller is expected to have
// already applied any required auth middleware on rg (Admin/Vendor read,
// Admin write).
func (c *CarController) Register(rg *gin.RouterGroup, adminOnly gin.HandlerFunc) {
	rg.GET("", c.list)
	rg.GET("/:id", c.getByID)
	rg.POST("", adminOnly, c.create)
	rg.PUT("/:id", adminOnly, c.update)
	rg.DELETE("/:id", adminOnly, c.delete)
}

func (c *CarController) list(ctx *gin.Context) {
	cars, err := c.svc.List(ctx.Request.Context())
	if err != nil {
		writeError(ctx, err)
		return
	}
	resp := make([]dto.CarResponse, len(cars))
	for i, car := range cars {
		resp[i] = dto.CarToResponse(car)
	}
	ctx.JSON(http.StatusOK, resp)
}

func (c *CarController) getByID(ctx *gin.Context) {
	id, ok := parseID(ctx)
	if !ok {
		return
	}
	car, err := c.svc.Get(ctx.Request.Context(), id)
	if err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, dto.CarToResponse(car))
}

func (c *CarController) create(ctx *gin.Context) {
	var req dto.CarRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	created, err := c.svc.Create(ctx.Request.Context(), req.ToDomain())
	if err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, dto.CarToResponse(created))
}

func (c *CarController) update(ctx *gin.Context) {
	id, ok := parseID(ctx)
	if !ok {
		return
	}
	var req dto.CarRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	updated, err := c.svc.Update(ctx.Request.Context(), id, req.ToDomain())
	if err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, dto.CarToResponse(updated))
}

func (c *CarController) delete(ctx *gin.Context) {
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
