package auth

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims is embedded in access JWTs.
type Claims struct {
	UserID  string   `json:"uid"`
	Email   string   `json:"email"`
	IsAdmin bool     `json:"adm"`
	TeamIDs []string `json:"teams,omitempty"`
	jwt.RegisteredClaims
}

// JWT holds signing configuration.
type JWT struct {
	secret        []byte
	accessTTL     time.Duration
	refreshMaxAge time.Duration // how long after exp we still allow refresh
}

// NewJWTFromEnv builds JWT helper using JWT_SECRET and optional JWT_ACCESS_TTL_MINUTES, JWT_REFRESH_MAX_AGE_HOURS.
func NewJWTFromEnv() (*JWT, error) {
	sec := os.Getenv("JWT_SECRET")
	if sec == "" {
		sec = "dev-only-change-me-32chars-min!!"
	}
	if len(sec) < 32 {
		return nil, errors.New("JWT_SECRET must be at least 32 characters")
	}
	ttl := 60 * time.Minute
	if v := os.Getenv("JWT_ACCESS_TTL_MINUTES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			ttl = time.Duration(n) * time.Minute
		}
	}
	refreshMax := 168 * time.Hour // 7d
	if v := os.Getenv("JWT_REFRESH_MAX_AGE_HOURS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			refreshMax = time.Duration(n) * time.Hour
		}
	}
	return &JWT{secret: []byte(sec), accessTTL: ttl, refreshMaxAge: refreshMax}, nil
}

// AccessTTL returns configured access token lifetime.
func (j *JWT) AccessTTL() time.Duration {
	return j.accessTTL
}

// IssueToken creates a signed access token with a new jti.
func (j *JWT) IssueToken(userID uuid.UUID, email string, isAdmin bool, teamIDs []string) (token string, jti string, err error) {
	jti = uuid.NewString()
	now := time.Now()
	claims := Claims{
		UserID:  userID.String(),
		Email:   email,
		IsAdmin: isAdmin,
		TeamIDs: teamIDs,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessTTL)),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err = t.SignedString(j.secret)
	return token, jti, err
}

// ValidateToken parses and validates signature + standard expiry.
func (j *JWT) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	t, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}
		return j.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !t.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// ParseForRefresh verifies signature without enforcing exp, then checks refresh window vs ExpiresAt.
func (j *JWT) ParseForRefresh(tokenString string) (*Claims, error) {
	claims := &Claims{}
	t, err := jwt.NewParser(jwt.WithoutClaimsValidation()).ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}
		return j.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !t.Valid {
		return nil, errors.New("invalid token signature")
	}
	if claims.ExpiresAt == nil {
		return nil, errors.New("token missing exp")
	}
	exp := claims.ExpiresAt.Time
	if time.Since(exp) > j.refreshMaxAge {
		return nil, errors.New("refresh window expired")
	}
	return claims, nil
}
