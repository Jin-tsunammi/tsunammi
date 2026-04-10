package auth

import (
	"errors"
	"mm/config"
	"mm/pkg/apperrors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/google/uuid"
)

type jwtAuthenticator struct {
	accessSignKey   string
	refreshSignKey  string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewJWTAuth(c *config.Config) JWTAuthenticator {
	return &jwtAuthenticator{
		accessSignKey:   c.Auth.SecretAccessSignKey,
		refreshSignKey:  c.Auth.SecretRefreshSignKey,
		accessTokenTTL:  c.Auth.AccessTokenTTL,
		refreshTokenTTL: c.Auth.RefreshTokenTTL,
	}
}

type TokenJWTClaims struct {
	TokenClaims
	jwt.RegisteredClaims
}

type TokenExpiration struct {
	Token     string
	ExpiresAt time.Time
}

func (a *jwtAuthenticator) GenerateTokenPair(claims TokenClaims) (*TokenPairWithExpiration, error) {
	refreshToken, err := a.generateToken(claims, a.refreshTokenTTL, a.refreshSignKey)
	if err != nil {
		return nil, err
	}

	accessToken, err := a.generateToken(claims, a.accessTokenTTL, a.accessSignKey)
	if err != nil {
		return nil, err
	}

	return &TokenPairWithExpiration{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (a *jwtAuthenticator) generateToken(
	claims TokenClaims,
	tokenLifetime time.Duration,
	signKey string,
) (*TokenExpiration, error) {
	if claims.UserID == uint64(0) {
		return nil, apperrors.Unauthorized(ErrInvalidTokenClaims)
	}

	mySigningKey := []byte(signKey)
	now := time.Now()
	expiresAt := now.Add(tokenLifetime)

	jwtClaims := TokenJWTClaims{
		TokenClaims: claims,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "some-app-api",
			Subject:   "client",
			ID:        claims.SessionID.String(),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims).SignedString(mySigningKey)
	if err != nil {
		return nil, apperrors.Unauthorized(ErrGenerateToken, err)
	}

	return &TokenExpiration{
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

func (a *jwtAuthenticator) ParseToken(authToken string, tokenType int) (*TokenClaims, error) {
	authToken = strings.TrimPrefix(authToken, "Bearer ")
	var keyFunc func(token *jwt.Token) (any, error)
	keyFunc = a.GetAccessKey
	if tokenType != 0 {
		keyFunc = a.GetRefreshKey
	}

	token, err := jwt.Parse(
		authToken,
		keyFunc,
		jwt.WithValidMethods([]string{"HS256"}))

	if err != nil {
		if errors.Is(err, jwt.ErrSignatureInvalid) || errors.Is(err, jwt.ErrTokenExpired) {
			return nil, apperrors.Unauthorized(ErrInvalidToken, err)
		}

		return nil, apperrors.Unauthorized("failed to parse jwt token", err)
	}

	if token == nil || !token.Valid {
		return nil, apperrors.Unauthorized(ErrInvalidToken)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, apperrors.Unauthorized(ErrInvalidTokenClaims)
	}

	return ExtractClaims(claims)
}

func ExtractClaims(claims jwt.MapClaims) (*TokenClaims, error) {
	userID, ok := extractID(claims, "user_id")
	if !ok {
		return nil, apperrors.Unauthorized("invalid user_id claim")
	}

	sessionID, ok := extractUUID(claims, "session_id")
	if !ok {
		return nil, apperrors.Unauthorized("invalid session_id claim")
	}

	parsed := &TokenClaims{
		UserID:    userID,
		SessionID: sessionID,
	}

	return parsed, nil
}

func ValueFromJWTClaims[T any](claims jwt.MapClaims, value string) (T, bool) {
	var zero T
	v, ok := claims[value]
	if !ok {
		return zero, false
	}

	typedValue, ok := v.(T)
	if !ok {
		return zero, false
	}

	return typedValue, true
}

func extractID(claims jwt.MapClaims, value string) (uint64, bool) {
	v, ok := claims[value]

	if !ok {
		return 0, false
	}

	id, ok := v.(float64)

	if !ok {
		return 0, false
	}

	return uint64(id), true
}

func extractUUID(claims jwt.MapClaims, value string) (uuid.UUID, bool) {
	v, ok := claims[value]
	if !ok {
		return uuid.Nil, false
	}

	stringUUID, ok := v.(string)
	if !ok {
		return uuid.Nil, false
	}

	parsed, err := uuid.Parse(stringUUID)
	if err != nil {
		return uuid.Nil, false
	}

	return parsed, true
}

func (a *jwtAuthenticator) GetAccessKey(token *jwt.Token) (any, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, apperrors.Unauthorized(ErrInvalidSigningMethod)
	}

	return []byte(a.accessSignKey), nil
}

func (a *jwtAuthenticator) GetRefreshKey(token *jwt.Token) (any, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, apperrors.Unauthorized(ErrInvalidSigningMethod)
	}

	return []byte(a.refreshSignKey), nil
}
