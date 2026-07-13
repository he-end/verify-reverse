package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/he-end/verify-reverse/verify/conf"
	"github.com/he-end/verify-reverse/verify/repository"
)

type JWTService struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

func NewJWTService(cfg conf.JWTConf) *JWTService {
	return &JWTService{
		accessSecret:  []byte(cfg.JWTAccessSecret),
		refreshSecret: []byte(cfg.JWTRefreshSecret),
		accessTTL:     cfg.JWTAccessTTL,
		refreshTTL:    cfg.JWTRefreshTTL,
	}
}

func (j *JWTService) AccessTTL() time.Duration {
	return j.accessTTL
}

func (j *JWTService) RefreshTTL() time.Duration {
	return j.refreshTTL
}

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type Claims struct {
	jwt.RegisteredClaims
	UserID   uuid.UUID `json:"uid"`
	UserHash string    `json:"uhash"`
}

func UserHash(passwordHash string) string {
	h := sha256.Sum256([]byte(passwordHash))
	return hex.EncodeToString(h[:])
}

func (j *JWTService) GenerateTokenPair(ctx context.Context, userID uuid.UUID, userHash string) (*TokenPair, error) {
	now := time.Now()
	accessExpires := now.Add(j.accessTTL)
	refreshExpires := now.Add(j.refreshTTL)

	accessTokenUUID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("generate token ID: %w", err)
	}
	accessTokenID := accessTokenUUID.String()

	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpires),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        accessTokenID,
		},
		UserID:   userID,
		UserHash: userHash,
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS512, claims).SignedString(j.accessSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	refreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    refreshExpires,
	}, nil
}

func (j *JWTService) ValidateAccessToken(ctx context.Context, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.accessSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", repository.ErrTokenInvalid, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, repository.ErrTokenInvalid
	}

	return claims, nil
}

func (j *JWTService) RefreshAccessToken(ctx context.Context, refreshToken, currentAccessToken string) (*TokenPair, error) {
	claims, err := j.ValidateAccessToken(ctx, currentAccessToken)
	if err != nil && !isTokenExpiredError(err) {
		return nil, fmt.Errorf("validate current token: %w", err)
	}

	return j.GenerateTokenPair(ctx, claims.UserID, claims.UserHash)
}

func isTokenExpiredError(err error) bool {
	for err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return true
		}
		err = errors.Unwrap(err)
	}
	return false
}

func generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}
	return hex.EncodeToString(b), nil
}
