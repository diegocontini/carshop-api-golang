package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"utfpr.edu.br/carshop-api/src/controller/dto"
	"utfpr.edu.br/carshop-api/src/domain"
	"utfpr.edu.br/carshop-api/src/service"
)

type AuthController struct {
	jwt   *service.JWTService
	users *service.UserService
}

func NewAuthController(jwt *service.JWTService, users *service.UserService) *AuthController {
	return &AuthController{jwt: jwt, users: users}
}

func (c *AuthController) Register(rg *gin.RouterGroup) {
	rg.POST("/token", c.token)
}

func (c *AuthController) token(ctx *gin.Context) {
	var req dto.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Username and password are required"})
		return
	}

	user, err := c.users.Authenticate(ctx.Request.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid username or password"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "internal error"})
		return
	}

	tok, exp, err := c.jwt.Generate(user.Username, user.Role)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "internal error"})
		return
	}

	ctx.JSON(http.StatusOK, dto.TokenResponse{
		Token:     tok,
		ExpiresAt: exp,
		TokenType: "Bearer",
	})
}
