package jwt

import (
	"fmt"
	"time"

	"realtime-api/internal/config"
	"realtime-api/internal/model"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTService struct {
	config *config.JWTConfig
}

type Claims struct {
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	DeviceID  string    `json:"device_id"`
	SessionID uuid.UUID `json:"session_id"`
	jwt.RegisteredClaims
}

var Service *JWTService

func Init(cfg *config.JWTConfig) *JWTService {
	service := &JWTService{
		config: cfg,
	}
	Service = service
	return service
}

func (j *JWTService) GenerateTokens(user *model.User, sessionID uuid.UUID, deviceID string) (string, string, time.Time, error) {
	// Access Token
	accessExpiry := time.Now().Add(time.Duration(j.config.AccessTokenTTL) * time.Minute)
	accessClaims := &Claims{
		UserID:    user.ID,
		Username:  user.Username,
		Email:     user.Email,
		DeviceID:  deviceID,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "realtime-api",
			Subject:   user.ID.String(),
			ID:        sessionID.String(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(j.config.SecretKey))
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Refresh Token
	refreshExpiry := time.Now().Add(time.Duration(j.config.RefreshTokenTTL) * time.Hour)
	refreshClaims := &Claims{
		UserID:    user.ID,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "realtime-api",
			Subject:   user.ID.String(),
			ID:        sessionID.String(),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(j.config.SecretKey))
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessTokenString, refreshTokenString, accessExpiry, nil
}

func (j *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.config.SecretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func (j *JWTService) RefreshAccessToken(refreshToken string) (string, time.Time, error) {
	claims, err := j.ValidateToken(refreshToken)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Generate new access token with same claims but new expiry
	accessExpiry := time.Now().Add(time.Duration(j.config.AccessTokenTTL) * time.Minute)
	newClaims := &Claims{
		UserID:    claims.UserID,
		Username:  claims.Username,
		Email:     claims.Email,
		DeviceID:  claims.DeviceID,
		SessionID: claims.SessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "realtime-api",
			Subject:   claims.UserID.String(),
			ID:        claims.SessionID.String(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	accessTokenString, err := accessToken.SignedString([]byte(j.config.SecretKey))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to generate new access token: %w", err)
	}

	return accessTokenString, accessExpiry, nil
}

func GetService() *JWTService {
	return Service
}
