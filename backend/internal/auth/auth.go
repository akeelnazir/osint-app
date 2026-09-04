// Package auth implements JWT issuance/verification and password hashing.
package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Claims is the JWT payload shared by access and refresh tokens.
type Claims struct {
	UserID int      `json:"uid"`
	Role   string   `json:"role"`
	Type   string   `json:"typ"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// Service handles token signing/verification and password hashing.
type Service struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

// New creates an auth Service.
func New(accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
	}
}

// HashPassword bcrypts a plaintext password.
func HashPassword(pw string) (string, error) {
	if len(pw) < 8 {
		return "", errors.New("password must be at least 8 characters")
	}
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(b), nil
}

// VerifyPassword checks a plaintext password against a bcrypt hash.
func VerifyPassword(hash, pw string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw))
}

// IssueAccessToken creates a signed access JWT for the user.
func (s *Service) IssueAccessToken(userID int, role string) (string, time.Time, error) {
	exp := time.Now().Add(s.accessTTL)
	claims := Claims{
		UserID: userID,
		Role:   role,
		Type:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString(s.accessSecret)
	return signed, exp, err
}

// IssueRefreshToken creates a signed refresh JWT for the user.
func (s *Service) IssueRefreshToken(userID int, role string) (string, time.Time, error) {
	exp := time.Now().Add(s.refreshTTL)
	claims := Claims{
		UserID: userID,
		Role:   role,
		Type:   "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString(s.refreshSecret)
	return signed, exp, err
}

// VerifyAccessToken parses and validates an access token.
func (s *Service) VerifyAccessToken(tokenStr string) (*Claims, error) {
	return s.verify(tokenStr, s.accessSecret, "access")
}

// VerifyRefreshToken parses and validates a refresh token.
func (s *Service) VerifyRefreshToken(tokenStr string) (*Claims, error) {
	return s.verify(tokenStr, s.refreshSecret, "refresh")
}

func (s *Service) verify(tokenStr string, secret []byte, wantType string) (*Claims, error) {
	claims := &Claims{}
	tok, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !tok.Valid {
		return nil, errors.New("invalid token")
	}
	if claims.Type != wantType {
		return nil, fmt.Errorf("expected %s token, got %s", wantType, claims.Type)
	}
	return claims, nil
}
