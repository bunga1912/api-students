package helper

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`

	jwt.RegisteredClaims
}

// JWT menyimpan secret dan masa berlaku token.
// Dibuat sekali di main.go, tidak membaca env/config sendiri.
type JWT struct {
	secret []byte
	ttl    time.Duration
}

func NewJWT(secret string, ttl time.Duration) (*JWT, error) {
	if secret == "" {
		return nil, errors.New("JWT_SECRET belum dikonfigurasi")
	}

	if ttl <= 0 {
		return nil, errors.New("masa berlaku token harus lebih dari 0")
	}

	return &JWT{
		secret: []byte(secret),
		ttl:    ttl,
	}, nil
}

func (j *JWT) Generate(
	userID int64,
	email string,
	role string,
) (string, error) {

	now := time.Now()

	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.ttl)),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString(j.secret)
}

func (j *JWT) Parse(tokenString string) (*Claims, error) {

	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {

			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("algoritma JWT tidak valid")
			}

			return j.secret, nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)

	if !ok || !token.Valid {
		return nil, errors.New("token tidak valid")
	}

	return claims, nil
}

func ExtractBearerToken(header string) (string, error) {

	parts := strings.SplitN(header, " ", 2)

	if len(parts) != 2 ||
		!strings.EqualFold(parts[0], "Bearer") {

		return "", fmt.Errorf(
			"format authorization harus Bearer <token>",
		)
	}

	if parts[1] == "" {
		return "", errors.New("token tidak ditemukan")
	}

	return parts[1], nil
}