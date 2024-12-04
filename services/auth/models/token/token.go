package token

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type Claim struct {
	Id    uuid.UUID
	Role  string
	Email string
	jwt.RegisteredClaims
}
