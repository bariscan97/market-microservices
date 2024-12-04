package model

import (
	"github.com/google/uuid"
)

type CustomerModel struct {
	CustomerId  uuid.UUID `bson:"customer_id,omitempty"`
	FirstName   string    `json:"first_name,omitempty" bson:"firstname,omitempty"`
	LastName    string    `json:"last_name,omitempty" bson:"lastname,omitempty"`
	Address     string    `json:"address,omitempty" bson:"address,omitempty"`
	Location    string    `json:"location,omitempty" bson:"location,omitempty"`
	PhoneNumber string    `json:"phone_number,omitempty" bson:"phone_number,omitempty"`
}
