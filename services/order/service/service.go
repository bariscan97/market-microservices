package service

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderService struct {
	pool *pgxpool.Pool
	
}

func NewService(pool *pgxpool.Pool) *OrderService {
	return &OrderService{
		pool: pool,
	}
}

func rollback(tx pgx.Tx, err *error) {
	if *err != nil {
		rollbackErr := tx.Rollback(context.TODO())
		if rollbackErr != nil {
			log.Printf("Rollback failed: %v", rollbackErr)
		} else {
			log.Println("Transaction rolled back successfully.")
		}
	}
}

type Order struct {
	ID         uuid.UUID
	CustomerID uuid.UUID
	Address    string
	Email      string
	TotalPrice float64
	Status     string
	Name       string
	ImageUrl   string
	Quantity   int
	Price      float64
}

func (orderService *OrderService) GetOrderById(id uuid.UUID) (*Order, error) {
	sql := `
		SELECT 
			o.id,
			o.customer_id,
			o.address,
			o.email,
			o.total_price,
			o.status,
			i.name,
			i.image,
			i.quantity,
			i.price 
		FROM 
			orders AS o
		LEFT JOIN 
			order_items AS i
		ON 
			o.id = i.order_id
		WHERE
			o.id = $1
	`
	var order Order
	if err := orderService.pool.QueryRow(context.Background(), sql, id).Scan(
		&order.ID,
		&order.CustomerID,
		&order.Address,
		&order.Email,
		&order.TotalPrice,
		&order.Status,
		&order.Name,
		&order.ImageUrl,
		&order.Quantity,
		&order.Price,
	); err != nil {
		return &Order{}, nil
	}

	return &order, nil
}

func (orderService *OrderService) GetAllOrders(customerID *uuid.UUID, page string) ([]*Order, error) {
	sql := `
		SELECT 
			o.id,
			o.customer_id,
			o.address,
			o.email,
			o.total_price,
			o.status,
			i.name,
			i.image,
			i.quantity,
			i.price 
		FROM 
			orders AS o
		LEFT JOIN 
			order_items AS i
		ON 
			o.id = i.order_id
	`
	var conditions string
	parameters := []any{page}
	if customerID != nil {
		conditions = " WHERE o.customer_id = $2 "
		parameters = append(parameters, customerID)
	}
	pagination := `
		ORDER BY 
			created_at DESC
		LIMIT 
			15 
    	OFFSET 
			$1 * 15
	`
	sql += conditions + pagination

	rows, err := orderService.pool.Query(context.Background(), sql, parameters...)
	if err != nil {
		return []*Order{}, nil
	}

	defer rows.Close()

	var orders []*Order

	for rows.Next() {
		var order Order
		if err := rows.Scan(
			&order.ID,
			&order.CustomerID,
			&order.Address,
			&order.Email,
			&order.TotalPrice,
			&order.Status,
			&order.Name,
			&order.ImageUrl,
			&order.Quantity,
			&order.Price,
		); err != nil {
			return []*Order{}, nil
		}
		orders = append(orders, &order)
	}
	return orders, nil
}

func (orderService *OrderService) CreateOrder(ctx context.Context, order map[string]interface{}) error {

	var (
		ID  uuid.UUID
		sql string
		err error
	)

	tx, err := orderService.pool.Begin(ctx)

	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer rollback(tx, &err)

	sql = `
		INSERT INTO orders(customer_id, address, email, total_price) VALUES($1, $2, $3, $4) RETURNING id
	`

	if err := orderService.pool.QueryRow(
		ctx,
		sql,
		order["customer_id"],
		order["address"],
		order["email"],
		order["total"],
	).Scan(&ID); err != nil {
		return err
	}

	items, ok := order["cart_items"].([]interface{})
	if ok {
		for _, val := range items {
			item, isMap := val.(map[string]interface{})
			if !isMap {
				continue
			}
			sql = `
				INSERT INTO order_items(order_id ,product_id, name, quantity, price, image) VALUES($1,$2,$3,$4,$5,$6)
			`
			if _, err := orderService.pool.Exec(
				ctx,
				sql,
				ID,
				item["id"],
				item["name"],
				item["quantity"],
				item["price"],
				item["image_url"],
			); err != nil {
				log.Printf("failed to order item: %s", err)
				continue
			}
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
