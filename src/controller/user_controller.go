package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"utfpr.edu.br/carshop-api/src/controller/dto"
	"utfpr.edu.br/carshop-api/src/domain"
	"utfpr.edu.br/carshop-api/src/service"
)

type UserController struct {
	svc *service.UserService
}

func NewUserController(svc *service.UserService) *UserController {
	return &UserController{svc: svc}
}

func (c *UserController) Register(rg *gin.RouterGroup) {
	rg.GET("", c.list)
	rg.GET("/:id", c.getByID)
	rg.POST("", c.create)
	rg.PUT("/:id", c.update)
	rg.DELETE("/:id", c.delete)
}

func (c *UserController) list(ctx *gin.Context) {
	users, err := c.svc.List(ctx.Request.Context())
	if err != nil {
		writeError(ctx, err)
		return
	}
	resp := make([]dto.UserResponse, len(users))
	for i, u := range users {
		resp[i] = dto.UserToResponse(u)
	}
	ctx.JSON(http.StatusOK, resp)
}

func (c *UserController) getByID(ctx *gin.Context) {
	id, ok := parseID(ctx)
	if !ok {
		return
	}
	u, err := c.svc.Get(ctx.Request.Context(), id)
	if err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, dto.UserToResponse(u))
}

func (c *UserController) create(ctx *gin.Context) {
	var req dto.UserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	created, err := c.svc.Create(ctx.Request.Context(), req.ToDomain())
	if err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, dto.UserToResponse(created))
}

func (c *UserController) update(ctx *gin.Context) {
	id, ok := parseID(ctx)
	if !ok {
		return
	}
	var req dto.UserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	updated, err := c.svc.Update(ctx.Request.Context(), id, req.ToDomain())
	if err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, dto.UserToResponse(updated))
}

func (c *UserController) delete(ctx *gin.Context) {
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

func parseID(ctx *gin.Context) (int64, bool) {
	raw := ctx.Param("id")
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid id"})
		return 0, false
	}
	return id, true
}

// writeError maps domain errors to HTTP responses.
func writeError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		ctx.JSON(http.StatusNotFound, gin.H{"message": "not found"})
	case errors.Is(err, domain.ErrConflict):
		ctx.JSON(http.StatusConflict, gin.H{"message": err.Error()})
	case errors.Is(err, domain.ErrInvalid):
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	case errors.Is(err, domain.ErrForbidden):
		ctx.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
	default:
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "internal error"})
	}
}
