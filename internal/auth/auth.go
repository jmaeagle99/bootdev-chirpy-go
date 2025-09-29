package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(password string) (string, error) {
	return argon2id.CreateHash(password, argon2id.DefaultParams)
}

func CheckPasswordHash(password, hash string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, hash)
}

func decodeTokenSecret(tokenSecret string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(tokenSecret)
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	decodedTokenSecret, err := decodeTokenSecret(tokenSecret)
	if err != nil {
		return "", err
	}

	now := time.Now().UTC()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
		Subject:   userID.String(),
	})

	return token.SignedString(decodedTokenSecret)
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := jwt.RegisteredClaims{}

	_, err := jwt.ParseWithClaims(
		tokenString,
		&claims,
		func(token *jwt.Token) (interface{}, error) {
			return decodeTokenSecret(tokenSecret)
		})
	if err != nil {
		return uuid.Nil, err
	}

	subject, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, err
	}

	return subject, nil
}

const bearerPrefix = "Bearer "

func GetBearerToken(headers http.Header) (string, error) {
	value := headers.Get("Authorization")
	if len(value) == 0 {
		return "", fmt.Errorf("authorization header is missing or empty")
	}
	if !strings.HasPrefix(value, bearerPrefix) {
		return "", fmt.Errorf("authorization header is not a bearer token")
	}
	return value[len(bearerPrefix):], nil
}

func MakeRefreshToken() (string, error) {
	token := make([]byte, 32)

	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(token), nil
}
