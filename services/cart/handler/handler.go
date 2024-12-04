package handler

import (
	// pb "cart-service/grpc/pb"
	"cart-service/models"
	"cart-service/service"
	"encoding/json"
	"net/http"
	
	"github.com/google/uuid"
	
)

type CartHandler struct {
	cartCache *service.RedisClient
	
}

func NewCustomerController(cartCache *service.RedisClient) *CartHandler {
	return &CartHandler{
		cartCache: cartCache,
	}
}

func (cartHandler *CartHandler) X(w http.ResponseWriter, r *http.Request) {
	ID, err := uuid.Parse(r.Header.Get("X-User-ID"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// defer func() {
	// 	cancel()
	// 	if rec := recover(); rec != nil {
	// 		log.Printf("Recovered from panic: %v", rec)
	// 		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	// 	}
	// }()
	claim := map[string]string{
		"customer_id" : r.Header.Get("X-User-Id"),
		"email" : r.Header.Get("X-User-email"),
	}
	if err := cartHandler.cartCache.PushToOrder(claim, ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return 
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "succesful",
	})
}

func (cartHandler *CartHandler) AddItemToCart(w http.ResponseWriter, r *http.Request) {
	ID, err := uuid.Parse(r.Header.Get("X-User-ID"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
    var item models.CartItem
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&item) ; err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_ , err = uuid.Parse(item.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	cart, err := cartHandler.cartCache.AddItemToCart(ID, item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return 
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cart)
}

func (cartHandler *CartHandler) GetCartByCustomerId(w http.ResponseWriter, r *http.Request) {
	ID, err := uuid.Parse(r.Header.Get("X-User-ID"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	cart, err := cartHandler.cartCache.GetCartByCustomerId(ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return 
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cart)
}
    
func (cartHandler *CartHandler) DeleteCartByCustomerId(w http.ResponseWriter, r *http.Request) {
	ID, err := uuid.Parse(r.Header.Get("X-User-ID"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := cartHandler.cartCache.ResetCartByCustomerId(ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return 
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"deleted": "successful",
	})
}



