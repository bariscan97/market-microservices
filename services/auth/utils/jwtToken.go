package utils

import (
	"auth-service/models/token"
	"os"
	"time"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

func GenerateJwtToken(id uuid.UUID, role string, email string) (string, error){
	
	expirationTime := time.Now().Add(72 * time.Hour)
	
	var jwtKey = []byte(os.Getenv("JWT_SECRET"))
	
	claims := &token.Claim{
		Id : id,
		Role: role,
		Email : email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}