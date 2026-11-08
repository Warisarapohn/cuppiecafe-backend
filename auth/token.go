package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

type Issuer struct {
	Secret []byte
	TTL    time.Duration
}

// Issue ออก Token ให้สมาชิกหนึ่งคน
func (i *Issuer) Issue(memberID string, role string) (string, error) {
	if len(i.Secret) == 0 {
		return "", errors.New("JWT secret is empty")
	}

	now := time.Now()

	claims := Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   memberID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(i.TTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(i.Secret)
}

// Parse ตรวจสอบ Token และคืน ID สมาชิกกับ Role
func (i *Issuer) Parse(tokenString string) (string, string, error) {
	var claims Claims

	_, err := jwt.ParseWithClaims(
		tokenString,
		&claims,
		func(t *jwt.Token) (any, error) {
			return i.Secret, nil
		},
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithExpirationRequired(),
	)

	if err != nil {
		return "", "", err
	}

	if claims.Subject == "" {
		return "", "", errors.New("bad subject")
	}

	return claims.Subject, claims.Role, nil
}
