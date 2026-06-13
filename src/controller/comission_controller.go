package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"utfpr.edu.br/carshop-api/src/controller/dto"
	"utfpr.edu.br/carshop-api/src/service"
)

type ComissionController struct {
	svc *service.ComissionService
}

func NewComissionController(svc *service.ComissionService) *ComissionController {
	return &ComissionController{svc: svc}
}

func (c *ComissionController) Register(rg *gin.RouterGroup, adminOnly gin.HandlerFunc) {
	rg.GET("", c.list)
	rg.GET("/:id", c.getByID)
	rg.POST("", adminOnly, c.create)
	rg.PUT("/:id", adminOnly, c.update)
	rg.DELETE("/:id", adminOnly, c.delete)
}

func (c *ComissionController) list(ctx *gin.Context) {
	var vendorID *int64
	if v := ctx.Query("vendorId"); v != "" {
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid vendorId"})
			return
		}
		vendorID = &parsed
	}
	rows, err := c.svc.List(ctx.Request.Context(), vendorID)
	if err != nil {
		writeError(ctx, err)
		return
	}
	resp := make([]dto.ComissionResponse, len(rows))
	for i, r := range rows {
		resp[i] = dto.ComissionToResponse(r)
	}
	ctx.JSON(http.StatusOK, resp)
}

func (c *ComissionController) getByID(ctx *gin.Context) {
	id, ok := parseID(ctx)
	if !ok {
		return
	}
	r, err := c.svc.Get(ctx.Request.Context(), id)
	if err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, dto.ComissionToResponse(r))
}

func (c *ComissionController) create(ctx *gin.Context) {
	var req dto.ComissionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	created, err := c.svc.Create(ctx.Request.Context(), req.ToDomain())
	if err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, dto.ComissionToResponse(created))
}

func (c *ComissionController) update(ctx *gin.Context) {
	id, ok := parseID(ctx)
	if !ok {
		return
	}
	var req dto.ComissionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	updated, err := c.svc.Update(ctx.Request.Context(), id, req.ToDomain())
	if err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, dto.ComissionToResponse(updated))
}

func (c *ComissionController) delete(ctx *gin.Context) {
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
