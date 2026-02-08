package auth

import (
	"errors"
	"time"

	"github.com/bsrodrigue/appshare-backend/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenType distinguishes between access and refresh tokens.
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Claims represents the JWT claims for our tokens.
type Claims struct {
	jwt.RegisteredClaims
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	TokenType TokenType `json:"token_type"`
}

// TokenPair contains both access and refresh tokens.
type TokenPair struct {
	AccessToken           string    `json:"access_token"`
	RefreshToken          string    `json:"refresh_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
	TokenType             string    `json:"token_type"` // Always "Bearer"
}

// JWTConfig holds JWT configuration.
type JWTConfig struct {
	SecretKey            string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	Issuer               string
}

// DefaultJWTConfig returns sensible defaults.
func DefaultJWTConfig(secretKey string) JWTConfig {
	return JWTConfig{
		SecretKey:            secretKey,
		AccessTokenDuration:  15 * time.Minute,   // Short-lived for security
		RefreshTokenDuration: 7 * 24 * time.Hour, // 7 days
		Issuer:               "appshare",
	}
}

// JWTService handles JWT token generation and validation.
type JWTService struct {
	config JWTConfig
}

// NewJWTService creates a new JWT service.
func NewJWTService(config JWTConfig) *JWTService {
	return &JWTService{config: config}
}

// GenerateTokenPair creates both access and refresh tokens for a user.
func (s *JWTService) GenerateTokenPair(user *domain.User) (*TokenPair, error) {
	now := time.Now()

	// Generate access token
	accessToken, accessExp, err := s.generateToken(user, AccessToken, now)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, refreshExp, err := s.generateToken(user, RefreshToken, now)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  accessExp,
		RefreshTokenExpiresAt: refreshExp,
		TokenType:             "Bearer",
	}, nil
}

// generateToken creates a single JWT token.
func (s *JWTService) generateToken(user *domain.User, tokenType TokenType, now time.Time) (string, time.Time, error) {
	var duration time.Duration
	if tokenType == AccessToken {
		duration = s.config.AccessTokenDuration
	} else {
		duration = s.config.RefreshTokenDuration
	}

	expiresAt := now.Add(duration)

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			Issuer:    s.config.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			ID:        uuid.NewString(), // Unique token ID for potential revocation
		},
		UserID:    user.ID,
		Email:     user.Email,
		TokenType: tokenType,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.config.SecretKey))
	if err != nil {
		return "", time.Time{}, err
	}

	return signedToken, expiresAt, nil
}

// ValidateAccessToken validates an access token and returns the claims.
func (s *JWTService) ValidateAccessToken(tokenString string) (*Claims, error) {
	claims, err := s.validateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != AccessToken {
		return nil, domain.NewAppError(domain.CodeTokenInvalid, "invalid token type: expected access token")
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token and returns the claims.
func (s *JWTService) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := s.validateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != RefreshToken {
		return nil, domain.NewAppError(domain.CodeTokenInvalid, "invalid token type: expected refresh token")
	}

	return claims, nil
}

// validateToken parses and validates a JWT token.
func (s *JWTService) validateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.config.SecretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, domain.ErrTokenExpired
		}
		return nil, domain.ErrTokenInvalid
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, domain.ErrTokenInvalid
	}

	return claims, nil
}

// RefreshTokens generates a new token pair from a valid refresh token.
// This is used when the access token expires but the refresh token is still valid.
func (s *JWTService) RefreshTokens(refreshTokenString string, user *domain.User) (*TokenPair, error) {
	// Validate the refresh token
	claims, err := s.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return nil, err
	}

	// Verify the user ID matches
	if claims.UserID != user.ID {
		return nil, domain.ErrTokenInvalid
	}

	// Generate new token pair
	return s.GenerateTokenPair(user)
}
