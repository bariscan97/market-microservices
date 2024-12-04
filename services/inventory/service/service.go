package service

import (
	model "inventory-service/models/product"
	"inventory-service/repository"
	"github.com/google/uuid"
)

type ProductService struct {
	productRepository repository.IProductRepository
}

type IProductService interface {
	CreateProduct(data *model.CreateProdcut) (uuid.UUID, error)
	GetAllProducts(category *string, page string) ([]model.FetchProduct, error)
	UpdateProductById(id uuid.UUID, fields map[string]interface{}) error
	GetCategories(page string) ([]map[string]interface{}, error)
	GetProductsByCategory(category string, page string) ([]model.FetchProduct, error)
	DeleteProductById(id uuid.UUID) error
	GetProductById(id uuid.UUID) (*model.FetchProduct, error)
}

func NewProductService(productRepository repository.IProductRepository) IProductService {
	return &ProductService{
		productRepository: productRepository,
	}
}
func (productService *ProductService) GetProductById(id uuid.UUID) (*model.FetchProduct, error) {
	return productService.productRepository.GetProductById(id)
}
func (productService *ProductService) GetCategories(page string) ([]map[string]interface{}, error) {
	return productService.productRepository.GetCategories(page)
}

func (productService *ProductService) GetProductsByCategory(category string, page string) ([]model.FetchProduct, error) {
	return productService.productRepository.GetProductsByCategory(category, page)
}

func (productService *ProductService) DeleteProductById(id uuid.UUID) error {
	return productService.productRepository.DeleteProductById(id)
}

func (productService *ProductService) UpdateProductById(id uuid.UUID, fields map[string]interface{}) error {
	return productService.productRepository.UpdateProductById(id, fields)
}

func (productService *ProductService) CreateProduct(data *model.CreateProdcut) (uuid.UUID, error) {
	return productService.productRepository.CreateProduct(data)
}

func (productService *ProductService) GetAllProducts(category *string, page string) ([]model.FetchProduct, error) {
	return productService.productRepository.GetAllProducts(category, page)
}

// func (productService *ProductService) PushProductToCatalog(ids []uuid.UUID) {
// 	for _, id := range ids {
// 		product, err := productService.GetProductById(id)
// 		if err != nil {
// 			log.Printf("error : %s", err)
// 			continue
// 		}
// 	}
	
// }