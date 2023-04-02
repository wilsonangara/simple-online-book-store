package auth

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
)

var (
	errIDIsRequired = errors.New("id is required")
	errParseClaims  = errors.New("parse claim error")

	ErrSecretIsRequired = errors.New("secret is required")
	ErrTokenExpired     = errors.New("token expired")
	ErrInvalidToken     = errors.New("invalid token")
)

//go:generate mockgen -source=auth.go -destination=mock/auth.go -package=mock
type AuthClient interface {
	// GenerateToken generates a valid authentication token.
	GenerateToken(id int64) (string, error)

	// ValidateToken recieves a signed token passed by the client validate it.
	ValidateToken(signedToken string) (int64, error)
}

type Client struct {
	secret string
}

// NewClient returns a wrapper around authentication client.
func NewClient(secret string) (*Client, error) {
	if secret == "" {
		return nil, ErrSecretIsRequired
	}

	return &Client{
		secret: secret,
	}, nil
}

// GenerateToken generates a valid authentication token.
func (c *Client) GenerateToken(id int64) (string, error) {
	if id == 0 {
		return "", errIDIsRequired
	}
	idStr := strconv.FormatInt(id, 10)

	currentTime := time.Now().UTC()

	// generate access token.
	tokenExpirationTime := currentTime.Add(24 * time.Hour)
	tokenClaims := jwt.StandardClaims{
		Subject:   idStr,
		ExpiresAt: tokenExpirationTime.Unix(),
		IssuedAt:  currentTime.Unix(),
		NotBefore: currentTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)
	tokenStr, err := token.SignedString([]byte(c.secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return tokenStr, nil
}

// ValidateToken recieves a signed token passed by the client validate it.
func (c *Client) ValidateToken(signedToken string) (int64, error) {
	token, err := jwt.ParseWithClaims(signedToken,
		&jwt.StandardClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return "", fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(c.secret), nil
		},
	)
	if err != nil {
		return 0, fmt.Errorf("failed while parsing token with claims: %w", err)
	}

	// assert jwt.MapClaims type
	claims, ok := token.Claims.(*jwt.StandardClaims)
	if !ok {
		return 0, errParseClaims
	}

	currentTime := time.Now().UTC().Unix()
	if ok := claims.VerifyExpiresAt(currentTime, true); !ok {
		return 0, ErrTokenExpired
	}
	if ok := claims.VerifyNotBefore(currentTime, true); !ok {
		return 0, ErrInvalidToken
	}

	// converts claims.Subject into id with type int.
	id, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to convert string subject to int id: %w", err)
	}

	return id, nil
}
