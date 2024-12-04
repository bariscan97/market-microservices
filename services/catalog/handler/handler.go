package handler

import (
	"catalog-service/cache"
	"catalog-service/utils"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
    "github.com/google/uuid"
	"github.com/gorilla/mux"
)

type ProductHandler struct {
	productCache *cache.RedisClient
}

func NewCustomerController (productCache *cache.RedisClient) *ProductHandler {
	return &ProductHandler{
		productCache: productCache,
	}
}

func (productHandler *ProductHandler) GetProductById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	parsedID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return 
	}
	product, err := productHandler.productCache.GetProductById(parsedID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return 
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(product)
}


func (productHandler *ProductHandler) GetGategories(w http.ResponseWriter, r *http.Request) {
	categories := productHandler.productCache.GetCategories()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(categories)
}

func (productHandler *ProductHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
	
	queryParams := r.URL.Query()

	search, err  := utils.DynamicSearch(queryParams)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return 
	}	
	
	page := queryParams.Get("page")
	
	num , err := strconv.Atoi(page)
	
	if err != nil {
		num = 0
	}	

	products, err := productHandler.productCache.GetProducts(context.Background(), search, num)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return 
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(products)
}
