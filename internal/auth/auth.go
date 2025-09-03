package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("Unable to hash password: %v", err)
	}
	return string(hashed), nil
}

func CheckPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return errors.New("Password doesn't match")
	}
	return nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:    "Chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userID.String(),
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := jwtToken.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", fmt.Errorf("Error trying to Sign jwtToken:%w", err)

	}
	return tokenString, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Signing Method does not match the Method used by the system to authenticate the token.")
		}
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("Unable to validate token: %w", err)
	}
	userId, err := token.Claims.GetSubject()
	return uuid.MustParse(userId), nil
}

func GetBearerToken(header http.Header) (string, error) {
	authtoken := header.Get("Authorization")
	if authtoken == "" {
		return "", fmt.Errorf("Auth Token not found in header.")
	}
	authtoken = strings.ReplaceAll(authtoken, "Bearer ", "")
	if authtoken == "" {
		return "", fmt.Errorf("Auth Token not found in header.")
	}

	return authtoken, nil
}

func GetAPIKey(header http.Header) (string, error) {
	apiKey := header.Get("Authorization")
	if apiKey == "" {
		return "", fmt.Errorf("Api key not found in header.")
	}
	apiKey = strings.ReplaceAll(apiKey, "ApiKey ", "")
	if apiKey == "" {
		return "", fmt.Errorf("Api key not found in header.")
	}
	return apiKey, nil
}

func MakeRefreshToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	refreshToken := hex.EncodeToString(b)
	return refreshToken
}
