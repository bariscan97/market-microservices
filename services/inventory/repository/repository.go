package repository

import (
	"context"
	model "inventory-service/models/product"
	"inventory-service/publisher"
	"inventory-service/utils"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"time"
	"encoding/json"
)

type ProductRepository struct {
	pool *pgxpool.Pool
}

type IProductRepository interface {
	CreateProduct(data *model.CreateProdcut) (uuid.UUID, error)
	GetAllProducts(category *string, page string) ([]model.FetchProduct, error)
	UpdateProductById(id uuid.UUID, fields map[string]interface{}) error
	GetCategories(page string) ([]map[string]interface{}, error)
	GetProductsByCategory(category string, page string) ([]model.FetchProduct, error)
	DeleteProductById(id uuid.UUID) error
	GetProductById(id uuid.UUID) (*model.FetchProduct, error)
	Listen(ctx context.Context) error
}

func NewProductRepo(pool *pgxpool.Pool) IProductRepository {
	return &ProductRepository{
		pool: pool,
	}
}

func (productRepository *ProductRepository) Listen(ctx context.Context) error {
	
	conn, err := productRepository.pool.Acquire(ctx)
	
	defer conn.Release()
	
	if err != nil {
		return err
	}

	trigers := []string{"product_update", "product_insert", "product_delete"}

	for _, triger := range trigers {
		str := fmt.Sprintf("LISTEN %v", triger)
		_, err = conn.Exec(ctx, str)
		if err != nil {
			return fmt.Errorf("failed to listen on %v: %v", triger, err)
		}
	}

	for {
			select {
			case <- ctx.Done():
				return ctx.Err()
			case <- time.After(30 * time.Second):	
				if err := productRepository.pool.Ping(ctx); err != nil {
					return err
				}
			default:
				notification, err := conn.Conn().WaitForNotification(ctx)
				if err != nil {
					log.Printf("notification error : %v\n", err)
					time.Sleep(time.Second)
					continue
				}
				
				var data map[string]interface{}

				if err := json.Unmarshal([]byte(notification.Payload), &data); err != nil {
					continue
				}

				payload := map[string]interface{}{
					"command" : notification.Channel,
					"payload" : data,
				}
				
				body, err := json.Marshal(payload)
				if err != nil {
					log.Printf("JSON error: %v\n", err)
					continue
				}
				if err := publisher.Publisher(body, ctx); err != nil {
					log.Printf("publisher error: %v\n", err)
				}
			}
			
		}
		
}

func (productRepository *ProductRepository) GetProductById(id uuid.UUID) (*model.FetchProduct, error) {
	
	sql := `
		SELECT id, name, slug, category, image_url, description, price, is_active, created_at, updated_at FROM products
		WHERE id = $1
	`
	
	ctx := context.Background()
	
	var product model.FetchProduct
	
	if err := productRepository.pool.QueryRow(ctx, sql, id).Scan(
		&product.Id,
		&product.Name,
		&product.Slug,
		&product.Image_url,
		&product.Category,
		&product.Description,
		&product.Price,
		&product.Is_active,
		&product.Created_at,
		&product.Updated_at,
	); err != nil {
		return &model.FetchProduct{}, err
	}

	return &product, nil
}


func (productRepository *ProductRepository) DeleteProductById(id uuid.UUID) error {
	
	sql := `
		DELETE FROM products
		WHERE id = $1
	`
	
	ctx := context.Background()

	result, err := productRepository.pool.Exec(ctx, sql, id)

	if err != nil {
		return fmt.Errorf("failed to delete product with id %v: %w", id, err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("no product found with id %v to delete", id)
	}
	
	return nil
}

func (productRepository *ProductRepository) GetProductsByCategory(category string, page string) ([]model.FetchProduct, error) {
	
	sql := `
		SELECT id, name, slug,image_url, category, description, price, is_active, created_at, updated_at FROM products
		WHERE category = $1
		ORDER BY created_at DESC
		LIMIT 15 
        OFFSET $2 * 15
	`
	
	ctx := context.Background()

	rows, err := productRepository.pool.Query(ctx, sql, category, page)

	if err != nil {
		return []model.FetchProduct{}, err
	}

	var products []model.FetchProduct
	
	for rows.Next() {
		var product model.FetchProduct
		if err := rows.Scan(
			&product.Id,
			&product.Name,
			&product.Slug,
			&product.Image_url,
			&product.Category,
			&product.Description,
			&product.Price,
			&product.Is_active,
			&product.Created_at,
			&product.Updated_at,
		); err != nil {
			return []model.FetchProduct{}, err
		}
		
		products = append(products, product)
	}
	
	return products, nil
}

func (productRepository *ProductRepository) GetCategories(page string) ([]map[string]interface{}, error) {
	
	ctx := context.Background()

	sql := `
		SELECT category, COUNT(*) as category_cnt FROM products
		GROUP BY category
		ORDER BY category_cnt DESC
		LIMIT 15 
        OFFSET $1 * 15
	`
	rows, err := productRepository.pool.Query(ctx, sql)

	if err != nil {
		return []map[string]interface{}{}, err
	}

	defer rows.Close()
	
	categories := []map[string]interface{}{}

	for rows.Next() {

		var (
			category_name   string
			category_count int
		)

		if err := rows.Scan(
			&category_name,
			&category_count,
		); err != nil {
			return []map[string]interface{}{}, err
		}

		category := map[string]interface{}{
			"category":   category_name,
			"category_count": category_count,
		}

		categories = append(categories, category)
	}

	return categories, nil
}
 
func (productRepository *ProductRepository) UpdateProductById(id uuid.UUID, fields map[string]interface{}) error {
	
	sql, parameters := utils.SqlUpdateQuery(
		fields,
		"products",
		map[string]interface{}{
			"id" : id,
		},
	)

	ctx := context.Background()
	
	result, err := productRepository.pool.Exec(ctx, sql, parameters...)

	if err != nil {
		return fmt.Errorf("failed to update product with id %v: %w", id, err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("no product found with id %v to update", id)
	}
	
	return nil

}


func (productRepository *ProductRepository) CreateProduct(data *model.CreateProdcut) (uuid.UUID, error) {

	ctx := context.Background()

	sql := `
		INSERT INTO products(name,slug,category,description,price,stock_quantity,image_url)
		VALUES($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	var ID uuid.UUID

	if err := productRepository.pool.QueryRow(ctx,
		sql, 
		data.Name, 
		data.Slug,
		data.Category, 
		data.Description, 
		data.Price,
		data.Stock_quantity, 
		data.Image_url,
		).Scan(&ID); err != nil {
			return uuid.UUID{}, err
		}

	return ID, nil
}

func (productRepository *ProductRepository) GetAllProducts(category *string, page string) ([]model.FetchProduct, error) {
	
	parameters := []interface{}{page}

	sql := "SELECT id, name, slug, image_url, category, description, price, is_active, created_at, updated_at FROM products"
	
	pagination := `
		ORDER BY created_at DESC
		LIMIT 15 
    	OFFSET $1 * 15
	`
	var (
		conditions string
		products []model.FetchProduct
	)

	if category != nil {
		parameters = append(parameters, *category)
		conditions += "	WHERE category = $2"
	}

	sql += conditions + pagination
	
	ctx := context.Background()

	rows, err := productRepository.pool.Query(ctx, sql, parameters...)

	if err != nil {
		return []model.FetchProduct{}, err
	}
	
	for rows.Next() {
		var product model.FetchProduct
		if err := rows.Scan(
			&product.Id,
			&product.Name,
			&product.Slug,
			&product.Image_url,
			&product.Category,
			&product.Description,
			&product.Price,
			&product.Is_active,
			&product.Created_at,
			&product.Updated_at,
		); err != nil {
			return []model.FetchProduct{}, err
		}
		
		products = append(products, product)
	}
	
	return products, nil
}

