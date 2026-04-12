package auth

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const defaultJWTSecret = "test-secret"

var (
	ErrTokenExpired = errors.New("token expired")
	ErrTokenInvalid = errors.New("token invalid")
)

// Claims represents validated token claims.
type Claims struct {
	UserID    string
	Role      string
	ExpiresAt time.Time
	IssuedAt  time.Time
}

// GenerateToken signs a JWT for the provided user identity and role.
func GenerateToken(userID, role string, ttl ...time.Duration) (string, error) {
	if strings.TrimSpace(userID) == "" {
		return "", errors.New("userID cannot be empty")
	}

	if strings.TrimSpace(role) == "" {
		role = "user"
	}

	tokenTTL := 24 * time.Hour
	if len(ttl) > 0 {
		tokenTTL = ttl[0]
	}

	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"sub":  userID,
		"role": role,
		"iat":  now.Unix(),
		"exp":  now.Add(tokenTTL).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(signingSecret()))
}

// ValidateToken validates a signed JWT and returns decoded claims.
func ValidateToken(tokenString string) (*Claims, error) {
	if strings.TrimSpace(tokenString) == "" {
		return nil, ErrTokenInvalid
	}

	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalid
		}
		return []byte(signingSecret()), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	if !parsedToken.Valid {
		return nil, ErrTokenInvalid
	}

	mapClaims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrTokenInvalid
	}

	userID, _ := mapClaims["sub"].(string)
	role, _ := mapClaims["role"].(string)
	expUnix, ok := unixFromClaim(mapClaims["exp"])
	if !ok {
		return nil, ErrTokenInvalid
	}
	issuedAtUnix, _ := unixFromClaim(mapClaims["iat"])

	if strings.TrimSpace(userID) == "" || strings.TrimSpace(role) == "" {
		return nil, ErrTokenInvalid
	}

	expiresAt := time.Unix(expUnix, 0).UTC()
	if time.Now().UTC().After(expiresAt) {
		return nil, ErrTokenExpired
	}

	return &Claims{
		UserID:    userID,
		Role:      role,
		ExpiresAt: expiresAt,
		IssuedAt:  time.Unix(issuedAtUnix, 0).UTC(),
	}, nil
}

func signingSecret() string {
	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if secret == "" {
		return defaultJWTSecret
	}

	return secret
}

func unixFromClaim(value any) (int64, bool) {
	switch v := value.(type) {
	case float64:
		return int64(v), true
	case int64:
		return v, true
	case int32:
		return int64(v), true
	case int:
		return int64(v), true
	default:
		return 0, false
	}
}
