package product

import (
	"time"
	"unicode"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type CreateProdcut struct {
	Name           string  `form:"name" validate:"required,isvalid"`
	Category       string  `form:"category" validate:"required,isvalid"`
	Description    string  `form:"description" validate:"required,isvalid"`
	Price          float64 `form:"price" validate:"required,gt=0"`
	Stock_quantity int64   `form:"stock_quantity" validate:"omitempty,gte=0"`
	Slug           string
	Image_url      string
}

type UpdateProduct struct {
	Name           string  `form:"name" validate:"omitempty,isvalid"`
	Category       string  `form:"category" validate:"omitempty,isvalid"`
	Description    string  `form:"description" validate:"omitempty,isvalid"`
	Price          float64 `form:"price" validate:"omitempty,gt=0"`
	Stock_quantity int64   `form:"stock_quantity" validate:"omitempty,gte=0"`
	Is_active      *bool    `form:"is_active" validate:"omitempty"`
	Slug           *string
	Image_url      *string
}

type FetchProduct struct {
	Id             uuid.UUID
	Name           string
	Slug           string
	Image_url      string
	Category       string
	Description    string
	Price          float64
	Stock_quantity int
	Is_active      bool
	Created_at     time.Time
	Updated_at     time.Time
}

func ValidName(fl validator.FieldLevel) bool {
	name := fl.Field().String()
	if name == "" {
		return false
	}
	for _, char := range name {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && !unicode.IsSpace(char) {
			return false
		}
	}
	return true
}
