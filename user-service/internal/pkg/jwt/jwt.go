package jwt

import (
	"fmt"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	jwtv5.RegisteredClaims
}

type Manager struct {
	secret []byte
	expire time.Duration
}

func NewManager(secret string, expireHours int) *Manager {
	if expireHours <= 0 {
		expireHours = 24
	}
	return &Manager{
		secret: []byte(secret),
		expire: time.Duration(expireHours) * time.Hour,
	}
}

func (m *Manager) Generate(userID uint64, username string) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(m.expire)

	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwtv5.RegisteredClaims{
			IssuedAt:  jwtv5.NewNumericDate(now),
			ExpiresAt: jwtv5.NewNumericDate(expiresAt),
		},
	}

	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return tokenString, expiresAt, nil
}

func (m *Manager) Parse(tokenString string) (*Claims, error) {
	token, err := jwtv5.ParseWithClaims(tokenString, &Claims{}, func(token *jwtv5.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwtv5.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
