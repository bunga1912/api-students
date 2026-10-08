package helper

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"siakad-mini/config"
)

type Claims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`

	jwt.RegisteredClaims
}

func GenerateJWT(
	userID int64,
	email string,
	role string,
) (string, error) {

	secret := config.GetEnv(
		"JWT_SECRET",
		"",
	)

	if secret == "" {
		return "", errors.New("JWT_SECRET belum dikonfigurasi")
	}

	now := time.Now()

	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:  strconv.FormatInt(userID, 10),
			IssuedAt: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(
				now.Add(24 * time.Hour),
			),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString(
		[]byte(secret),
	)
}

func ParseJWT(
	tokenString string,
) (*Claims, error) {

	secret := config.GetEnv(
		"JWT_SECRET",
		"",
	)

	if secret == "" {
		return nil, errors.New("JWT_SECRET belum dikonfigurasi")
	}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {

			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.New(
					"algoritma JWT tidak valid",
				)
			}

			return []byte(secret), nil
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

func ExtractBearerToken(
	header string,
) (string, error) {

	parts := strings.SplitN(
		header,
		" ",
		2,
	)

	if len(parts) != 2 ||
		!strings.EqualFold(parts[0], "Bearer") {

		return "", fmt.Errorf(
			"format authorization harus Bearer <token>",
		)
	}

	if parts[1] == "" {
		return "", errors.New(
			"token tidak ditemukan",
		)
	}

	return parts[1], nil
}
