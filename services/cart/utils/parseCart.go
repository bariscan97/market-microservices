package utils

import (
	"cart-service/models"	
	"encoding/json"
)

func PushRedis(cart models.Cart) ([]byte, error) {
	cartJSON, err := json.Marshal(cart)
	if err != nil {
		return []byte{}, err
	}
	encrypted, err := Encrypt(cartJSON)
	if err != nil {
		return []byte{}, err
	}
	return encrypted, nil
}

func ParseCart(payload []byte, cart *models.Cart) (*models.Cart, error) {
	JsonCart, err := Decrypt(payload)
	if err != nil {
		return &models.Cart{}, err
	}
	err = json.Unmarshal(JsonCart, &cart)
	if err != nil {
		return &models.Cart{}, err
	}
	return cart, nil
}