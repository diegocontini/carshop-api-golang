package unit

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"utfpr.edu.br/carshop-api/src/domain"
)

func TestUserRoleIsValid(t *testing.T) {
	require.True(t, domain.RoleAdmin.IsValid())
	require.True(t, domain.RoleVendor.IsValid())
	require.False(t, domain.UserRole("").IsValid())
	require.False(t, domain.UserRole("ADMIN").IsValid())
	require.False(t, domain.UserRole("guest").IsValid())
}

func TestUserCalcCommission(t *testing.T) {
	pct3 := int16(3)
	pct0 := int16(0)
	pctNeg := int16(-1)

	cases := []struct {
		name  string
		pct   *int16
		total string
		want  string
	}{
		{"nil percentage -> zero", nil, "100000.00", "0"},
		{"zero percentage -> zero", &pct0, "100000.00", "0"},
		{"negative percentage -> zero", &pctNeg, "100000.00", "0"},
		{"3 percent of 100k", &pct3, "100000.00", "3000"},
		{"3 percent of fractional", &pct3, "12345.67", "370.3701"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			u := domain.User{ComissionPerSaleInPercent: tc.pct}
			total, _ := decimal.NewFromString(tc.total)
			got := u.CalcCommission(total)
			want, _ := decimal.NewFromString(tc.want)
			require.True(t, got.Equal(want), "want %s, got %s", want, got)
		})
	}
}
