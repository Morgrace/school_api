package utils

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

var jwtKey []byte

// Initialize the key from environment variables
func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system env")
	}
	key := os.Getenv("JWT_SECRET_KEY")
	if key == "" {
		// As an AppSec engineer, never let the app run with a default or empty key
		panic("JWT_SECRET_KEY environment variable is not set")
	}
	jwtKey = []byte(key)
}

type CustomClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateJWT(userID string, role string) (string, error) {
	claims := CustomClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			// OWASP recommends short-lived access tokens
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "school-app", // Identify who created the token
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func ValidateJWT(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// AppSec Check: Ensure the algorithm is HMAC.
		// This prevents the "alg: none" attack where users bypass auth.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, errors.New("could not parse claims")
	}

	return claims, nil
}
