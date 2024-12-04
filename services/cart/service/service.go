package service

import (
	customer_pb "cart-service/grpc/pb/customer-rpc"
	product_pb "cart-service/grpc/pb/product-rpc"
	"cart-service/models"
	"cart-service/utils"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"golang.org/x/sync/semaphore"
)

type RedisClient struct {
	rdb             *redis.Client
	UserEvent       chan *uuid.UUID
	ProductEvents   chan map[string]interface{}
	NotifyQue       chan<- map[string]interface{}
	PushOrderQue    chan<- map[string]interface{}
	CustomerDialler customer_pb.CustomerServiceClient
	ProductDialler  product_pb.ProductServiceClient
}

func NewCacheClient(
	redisCli *redis.Client,
	notifyQue chan<- map[string]interface{},
	pushOrderQue chan<- map[string]interface{},
	customerClient customer_pb.CustomerServiceClient,
	productClient product_pb.ProductServiceClient,
) *RedisClient {
	return &RedisClient{
		rdb:             redisCli,
		UserEvent:       make(chan *uuid.UUID),
		ProductEvents:   make(chan map[string]interface{}, 10),
		NotifyQue:       notifyQue,
		PushOrderQue:    pushOrderQue,
		CustomerDialler: customerClient,
		ProductDialler:  productClient,
	}
}

func (redisClient *RedisClient) PushToOrder(claim map[string]string, customerID uuid.UUID) error {
	cart, err := redisClient.GetCartByCustomerId(customerID)
	if err != nil {
		return err
	}
	var total int
	order := make(map[string]interface{})
	cartItems := make([]*models.CartItem, 0)
	for _, item := range cart.CartItems {
		total += item.Quantity * int(item.Price)
		cartItems = append(cartItems, item)
	}

	customer, err := redisClient.CustomerDialler.GetCustomer(context.Background(), &customer_pb.CustomerReq{
		Id: customerID.String(),
	})
	fmt.Println("Customerr: ", customer)
	if err != nil {
		return err
	}
	
	product, err := redisClient.GetCartByCustomerId(customerID)
	if err != nil {
		return err
	}
	var ids []string
	for _, item := range product.CartItems {
		ids = append(ids, item.ID)
	}
	result, err := redisClient.ProductDialler.AllItemsExists(context.Background(),&product_pb.ProductReq{
		Id: ids,
	})
	if err != nil {
		return err
	}

	if !result.AllExist {
		return fmt.Errorf("some products not exist")
	}

	order["cart_items"] = cartItems
	order["total"] = total
	order["customer_id"] = claim["customer_id"]
	order["email"] = claim["email"]
	order["address"] = customer.Address

	redisClient.PushOrderQue <- order

	return nil
}

func (redisClient *RedisClient) AddItemToCart(customerID uuid.UUID, item models.CartItem) (models.Cart, error) {
	ctx := context.Background()
	key := fmt.Sprintf("cart:%v", customerID)
	exists, err := redisClient.rdb.Exists(context.Background(), key).Result()
	if err != nil {

		return models.Cart{}, err
	}
	setKey := fmt.Sprintf("productSet:%s", item.ID)
	var cart models.Cart
	if exists == 0 {
		if item.Quantity <= 0 {
			return models.Cart{}, fmt.Errorf("not valid count")
		}
		redisClient.rdb.SAdd(ctx, setKey, customerID.String())
		cart = models.Cart{
			CustomerID: customerID,
			CartItems:  make(map[string]*models.CartItem),
		}
		cart.CartItems[item.ID] = &item
	} else {
		encryptedFromRedis, err := redisClient.rdb.Get(ctx, key).Bytes()
		if err != nil {
			return models.Cart{}, err
		}
		JsonCart, err := utils.Decrypt(encryptedFromRedis)
		if err != nil {
			return models.Cart{}, err
		}
		err = json.Unmarshal(JsonCart, &cart)
		if err != nil {
			log.Fatalf("JSON unmarshalling hatasi: %v", err)
		}
		cartItem, ok := cart.CartItems[item.ID]
		if ok {
			newCount := cartItem.Quantity + item.Quantity
			if newCount <= 0 {
				redisClient.rdb.SRem(ctx, setKey, customerID.String())
				delete(cart.CartItems, item.ID)

				if len(cart.CartItems) == 0 {
					redisClient.rdb.Del(ctx, key)
					return models.Cart{}, nil
				}

			} else {
				cartItem.Quantity = newCount
			}

		} else {
			if item.Quantity >= 0 {
				cart.CartItems[item.ID] = &item
				redisClient.rdb.SAdd(ctx, setKey, customerID.String())
			} else {
				return models.Cart{}, fmt.Errorf("invalid count")
			}
		}
	}
	cartJSON, err := json.Marshal(cart)
	if err != nil {
		log.Fatalf("JSON marshalling error: %v", err)
	}
	encrypted, err := utils.Encrypt(cartJSON)
	if err != nil {
		return models.Cart{}, err
	}
	if err := redisClient.rdb.Set(ctx, key, encrypted, 0); err != nil {
		log.Printf("Redis set error: %v", err)
	}
	return cart, nil
}

func (redisClient *RedisClient) GetCartByCustomerId(customerID uuid.UUID) (models.Cart, error) {
	key := fmt.Sprintf("cart:%v", customerID)
	encryptedFromRedis, err := redisClient.rdb.Get(context.Background(), key).Bytes()
	if err != nil {
		return models.Cart{}, err
	}
	decryptedData, err := utils.Decrypt(encryptedFromRedis)
	if err != nil {
		return models.Cart{}, err
	}
	var cart models.Cart
	if err := json.Unmarshal(decryptedData, &cart); err != nil {
		log.Fatalf("JSON unmarshalling error: %v", err)
	}
	return cart, err
}

func (redisClient *RedisClient) ResetCartByCustomerId(customerID uuid.UUID) error {
	key := fmt.Sprintf("cart:%v", customerID)
	ctx := context.Background()
	redisClient.UserEvent <- &customerID
	deleted, err := redisClient.rdb.Del(ctx, key).Result()
	if err != nil || deleted == 0 {
		return err
	}
	return nil
}

func (redisClient *RedisClient) EventListener(ctx context.Context) error {

	ticker := time.NewTicker(30 * time.Minute)

	defer func() {
		ticker.Stop()
		close(redisClient.ProductEvents)
		close(redisClient.UserEvent)
		
	}()

loop:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := redisClient.rdb.Ping(context.Background()).Err(); err != nil {
				log.Printf("Error:  %v", err)
				break loop
			}
		case customerID, ok := <-redisClient.UserEvent:
			if !ok {
				continue
			}
			key := fmt.Sprintf("cart:%s", customerID.String())
			encryptedFromRedis, _ := redisClient.rdb.Get(context.Background(), key).Bytes()
			decryptedData, _ := utils.Decrypt(encryptedFromRedis)
			var (
				cart   models.Cart
				setKey string
			)
			if err := json.Unmarshal(decryptedData, &cart); err != nil {
				log.Printf("JSON unmarshalling error: %v", err)
			}
			_map := utils.StructToMap(cart.CartItems)
			for k := range _map {
				setKey = fmt.Sprintf("productSet:%s", k)
				redisClient.rdb.SRem(ctx, setKey, customerID.String())
			}
			redisClient.rdb.Del(ctx, key)
		case data, ok := <-redisClient.ProductEvents:
			if !ok {
				continue
			}
			event, _ := data["command"].(string)
			payload, _ := data["payload"].(map[string]interface{})
			ID, err := uuid.Parse(payload["id"].(string))
			if err != nil {
				log.Println("err", err.Error())
			}

			var (
				cursor uint64
				setKey string
			)

			setKey = fmt.Sprintf("productSet:%s", ID.String())

			for {
				members, cursor, err := redisClient.rdb.SScan(ctx, setKey, cursor, "*", 10).Result()
				if err != nil {
					continue
				}
				sem := semaphore.NewWeighted(10)
				var wg sync.WaitGroup
				for _, member := range members {
					wg.Add(1)
					if err := sem.Acquire(ctx, 1); err != nil {
						return err
					}
					go func(user string) {
						key := fmt.Sprintf("cart:%v", member)
						encryptedFromRedis, _ := redisClient.rdb.Get(ctx, key).Bytes()
						decryptedData, _ := utils.Decrypt(encryptedFromRedis)
						var cart models.Cart
						customerID, _ := uuid.Parse(user)
						if err := json.Unmarshal(decryptedData, &cart); err != nil {
							log.Printf("JSON unmarshalling error: %v", err)
						}
						defer func() {
							sem.Release(1)
							wg.Done()
							if len(cart.CartItems) == 0 {
								redisClient.rdb.Del(ctx, key)
								return
							}
							cartJSON, _ := json.Marshal(cart)
							encrypted, _ := utils.Encrypt(cartJSON)
							if err := redisClient.rdb.Set(ctx, key, encrypted, 0); err != nil {
								log.Printf("Redis set error: %v", err)
							}
						}()
						_, err = uuid.Parse(payload["id"].(string))
						if err != nil {

						}
						ItemID := payload["id"].(string)
						switch strings.Split(event, "_")[1] {
						case "update":
							for k, v := range payload {
								if k == "id" {
									continue
								}
								switch k {
								case "name":
									cart.CartItems[ItemID].Name = v.(string)
								case "category":
									cart.CartItems[ItemID].Category = v.(string)
								case "slug":
									cart.CartItems[ItemID].Slug = v.(string)
								case "image_url":
									cart.CartItems[ItemID].ImageUrl = v.(string)
								case "price":
									cart.CartItems[ItemID].Price = v.(float64)
								}
							}
							redisClient.NotifyQue <- map[string]interface{}{
								"customer_id": customerID,
								"content":     "updated some products from your card",
							}
						case "delete":
							delete(cart.CartItems, ItemID)

							redisClient.rdb.SRem(ctx, setKey, customerID.String())

							redisClient.NotifyQue <- map[string]interface{}{
								"customer_id": customerID,
								"content":     "removed some products from your card",
							}
						}
					}(member)
				}
				wg.Wait()
				setLength, _ := redisClient.rdb.SCard(ctx, setKey).Result()
				
				if setLength == 0 {
					redisClient.rdb.Del(ctx, setKey)
				}
				if cursor == 0 {
					break
				}
			}
		}
	}
	panic("somethin wrong")
}
