package controller

import (
	"customer-service/model"
	"customer-service/service"
	"encoding/json"
	"net/http"
	"github.com/google/uuid"
)

type CustomerController struct {
	customerService service.ICustomerService
}

type ICustomerController interface {

	GetCustomerById(w http.ResponseWriter, r *http.Request)
	DeleteCustomerById(w http.ResponseWriter, r *http.Request)
	AddCustomerFieldById(w http.ResponseWriter, r *http.Request)
}

func NewCustomerController (customerService service.ICustomerService) ICustomerController {
	return &CustomerController{
		customerService: customerService,
	}
}


func (customerController *CustomerController) GetCustomerById(w http.ResponseWriter, r *http.Request) {
	ID, err := uuid.Parse(r.Header.Get("X-User-ID"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user, err := customerController.customerService.GetCustomerById(ID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return 
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (customerController *CustomerController) DeleteCustomerById(w http.ResponseWriter, r *http.Request) {
	ID, err := uuid.Parse(r.Header.Get("X-User-ID"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := customerController.customerService.DeleteCustomerById(ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return 
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message" : "ok",
	})
}

func (customerController *CustomerController) AddCustomerFieldById(w http.ResponseWriter, r *http.Request) {
	var user model.CustomerModel
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return 
	}
	ID, err := uuid.Parse(r.Header.Get("X-User-ID"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user.CustomerId = ID
	if err := customerController.customerService.AddCustomerFieldById(ID, &user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return 
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}