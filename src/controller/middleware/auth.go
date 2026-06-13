package middleware

import (
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"

	"utfpr.edu.br/carshop-api/src/domain"
	"utfpr.edu.br/carshop-api/src/service"
)

const (
	ctxSubject = "auth.subject"
	ctxRole    = "auth.role"
)

// RequireAuth validates the Bearer token and stashes subject + role in the
// gin context. Without a valid token the request is short-circuited 401.
func RequireAuth(jwt *service.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		const prefix = "Bearer "
		if !strings.HasPrefix(header, prefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "missing bearer token"})
			return
		}
		claims, err := jwt.Parse(strings.TrimPrefix(header, prefix))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
			return
		}
		c.Set(ctxSubject, claims.Subject)
		c.Set(ctxRole, claims.Role)
		c.Next()
	}
}

// RequireRoles gates a route by role. Must run after RequireAuth.
// Roles are matched case-sensitive against the lowercase domain values.
func RequireRoles(allowed ...domain.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, ok := c.Get(ctxRole)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "no role on request"})
			return
		}
		role, ok := raw.(domain.UserRole)
		if !ok || !slices.Contains(allowed, role) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "forbidden"})
			return
		}
		c.Next()
	}
}

// Subject returns the authenticated username if present.
func Subject(c *gin.Context) (string, bool) {
	v, ok := c.Get(ctxSubject)
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

// Role returns the authenticated role if present.
func Role(c *gin.Context) (domain.UserRole, bool) {
	v, ok := c.Get(ctxRole)
	if !ok {
		return "", false
	}
	r, ok := v.(domain.UserRole)
	return r, ok
}
