package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"utfpr.edu.br/carshop-api/src/domain"
)

// JWTService signs and verifies HS256 bearer tokens for the API.
//
// Token shape mirrors the C# CarShopApi/src/Services/JwtService.cs:
//   - sub: username
//   - role: "admin" or "vendor"
//   - jti: random uuid for traceability
//   - iss, aud, exp: from Settings
type JWTService struct {
	secret     []byte
	issuer     string
	audience   string
	expiryMins int
}

func NewJWTService(secret, issuer, audience string, expiryMins int) *JWTService {
	return &JWTService{
		secret:     []byte(secret),
		issuer:     issuer,
		audience:   audience,
		expiryMins: expiryMins,
	}
}

type Claims struct {
	Role domain.UserRole `json:"role"`
	jwt.RegisteredClaims
}

// Generate produces a signed token plus its absolute expiration time.
func (s *JWTService) Generate(subject string, role domain.UserRole) (string, time.Time, error) {
	expiresAt := time.Now().Add(time.Duration(s.expiryMins) * time.Minute)
	claims := Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			Issuer:    s.issuer,
			Audience:  jwt.ClaimStrings{s.audience},
			ID:        uuid.NewString(),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign token: %w", err)
	}
	return signed, expiresAt, nil
}

var ErrInvalidToken = errors.New("invalid token")

// Parse validates a raw token string and returns the embedded claims.
func (s *JWTService) Parse(raw string) (*Claims, error) {
	claims := &Claims{}
	tok, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.secret, nil
	},
		jwt.WithIssuer(s.issuer),
		jwt.WithAudience(s.audience),
		jwt.WithValidMethods([]string{"HS256"}),
	)
	if err != nil || !tok.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
