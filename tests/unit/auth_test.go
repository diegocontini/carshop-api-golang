package unit

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"utfpr.edu.br/carshop-api/src/domain"
	"utfpr.edu.br/carshop-api/src/service"
)

const testSecret = "test-secret-that-is-at-least-32-chars-long"

func TestJWTRoundTrip(t *testing.T) {
	j := service.NewJWTService(testSecret, "CarShopApi", "CarShopApiClients", 60)
	tok, exp, err := j.Generate("alice", domain.RoleAdmin)
	require.NoError(t, err)
	require.True(t, strings.Count(tok, ".") == 2, "expected JWT format")
	require.True(t, exp.After(exp.Add(-1)))

	claims, err := j.Parse(tok)
	require.NoError(t, err)
	require.Equal(t, "alice", claims.Subject)
	require.Equal(t, domain.RoleAdmin, claims.Role)
}

func TestJWTRejectsWrongSecret(t *testing.T) {
	j := service.NewJWTService(testSecret, "iss", "aud", 60)
	tok, _, err := j.Generate("alice", domain.RoleVendor)
	require.NoError(t, err)

	other := service.NewJWTService("another-secret-that-is-at-least-32-bytes!!", "iss", "aud", 60)
	_, err = other.Parse(tok)
	require.ErrorIs(t, err, service.ErrInvalidToken)
}

func TestPasswordHashRoundTrip(t *testing.T) {
	h, err := service.HashPassword("hunter2")
	require.NoError(t, err)
	require.NotEqual(t, "hunter2", h)
	require.True(t, service.CheckPassword(h, "hunter2"))
	require.False(t, service.CheckPassword(h, "wrong"))
}
