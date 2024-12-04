package cache

import (
	"catalog-service/utils"
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type RedisClient struct {
	rdb    *redis.Client
	Add    chan map[string]interface{}
	Pop    chan map[string]interface{}
	Update chan map[string]interface{}
}

func NewCacheClient(redisCli *redis.Client) *RedisClient {
	return &RedisClient{
		rdb:    redisCli,
		Add:    make(chan map[string]interface{}, 10),
		Pop:    make(chan map[string]interface{}, 10),
		Update: make(chan map[string]interface{}, 10),
	}
}

type FetchProduct struct {
	Name          string    `json:"name"`
	Slug          string    `json:"slug"`
	ImageUrl      string    `json:"image_url"`
	Category      string    `json:"category"`
	Description   string    `json:"description"`
	Price         float64   `json:"price"`
	StockQuantity int       `json:"stock_quantity"`
	// IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
}

type Cat struct{
	Category string
	Count    int
}

func (redisCli *RedisClient) GetCategories() []Cat {
	
	ctx := context.Background()

	var (
		cursor uint64
		keys []Cat
		err error
	)
	
	prefix := "category:*"
	
	for {
		var batch []string
		batch, cursor, err = redisCli.rdb.Scan(ctx, cursor, prefix, 10).Result()
		if err != nil {
			log.Fatalf("Redis SCAN hatasi: %v", err)
		}
		for _ ,key := range batch {
			catName := strings.Split(key, ":")[1]
			
			val, err := redisCli.rdb.Get(ctx, key).Result()
			if err != nil {
				log.Println(err)
				continue
			}
			num, err := strconv.Atoi(val)
			if err != nil {
				log.Println(err)
				continue
			}
			keys = append(keys,
				Cat{
					Category: catName,
					Count: num ,
				},
			)
		}
		if cursor == 0 {
			break
		}
	}
	return keys
}

func (redisCli *RedisClient) GetProductById(productID uuid.UUID) (*FetchProduct, error) {
	ctx := context.Background()
	key := fmt.Sprintf("product:%s", productID.String())
	values, err  := redisCli.rdb.HGetAll(ctx, key).Result()
	if err == redis.Nil || err != nil {
		return &FetchProduct{}, err 
	}
	var product FetchProduct
	for key, val := range values {
		switch key {
		case "name":
			product.Name = val
		case "slug":
			product.Slug = val
		case "image_url":
			product.ImageUrl = val
		case "category":
			product.Category = val
		case "description":
			product.Description = val
		case "price":
			price, err := strconv.ParseFloat(val, 64)
			if err != nil {
				log.Printf("error: %s", err)
				break
			}
			product.Price = price
		case "stock_quantity":
			cnt, err := strconv.Atoi(val)
			if err != nil {
				log.Printf("error: %s", err)
				break
			}
			product.StockQuantity = cnt
		case "created_at":
			createdAt, err := utils.ConvertUnixToTime(val)
			if err != nil {
				log.Printf("error: %s", err)
				continue
			}
			product.CreatedAt = createdAt
		}
	}

	return &product, nil
	
}

func (redisCli *RedisClient) UpdateProduct(ctx context.Context, productID uuid.UUID, fields map[string]interface{}) error {
    key := fmt.Sprintf("product:%s", productID.String())
	str, ok := fields["is_active"].(string)
	if ok {
		boolValue, err := strconv.ParseBool(str)
		if err == nil {
			if !boolValue {
				redisCli.PopProduct(ctx, productID)
				// product, err := redisCli.GetProductById(productID)
				// if err != nil {
				// 	return nil
				// }
				// catKey := fmt.Sprintf("category:%s",product.Category)
				// val, check := redisCli.rdb.Get(ctx, catKey).Result()
				// if check == redis.Nil {
				// 	redisCli.rdb.Set(ctx, catKey, 1, 0)
				// }else {
				// 	cnt , err := strconv.Atoi(val)
				// 	if err == nil {
				// 		redisCli.rdb.Set(ctx, catKey, cnt - 1, 0)
				// 	}
				// }
				// if err := redisCli.rdb.Del(ctx,key).Err(); err != nil {
				// 	return err
				// }
				return nil
			}
		}
	}
    _, err := redisCli.rdb.HSet(ctx, key, fields).Result()
    if err != nil {
        return err
    }
    return nil
}

func (redisCli *RedisClient) PopProduct(ctx context.Context, productID uuid.UUID) error {
	product, err := redisCli.GetProductById(productID)
	if err == nil {
		catKey := fmt.Sprintf("category:%s",product.Category)
		val, check := redisCli.rdb.Get(ctx, catKey).Result()
		if check != redis.Nil {
			redisCli.rdb.Set(ctx, catKey, 1, 0)
			cnt , err := strconv.Atoi(val)
			if err == nil {
				if cnt > 1 {
					redisCli.rdb.Set(ctx, catKey, cnt - 1, 0)
				}else {
					redisCli.rdb.Del(ctx, catKey)
				}
			}
		}
	}
	key := fmt.Sprintf("product:%s", productID.String())
	if	err := redisCli.rdb.Del(ctx, key).Err(); err != nil {
		return err	
	}
	return nil
}

func (redisCli *RedisClient) AddProduct(ctx context.Context, data map[string]interface{}) error {
	id, _ := data["id"].(string)
	ID, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	catName := data["category"].(string)
	catKey := fmt.Sprintf("category:%s",catName)
	val, check := redisCli.rdb.Get(ctx, catKey).Result()
	if check == redis.Nil {
		redisCli.rdb.Set(ctx, catKey, 1, 0)
	}else {
		cnt , err := strconv.Atoi(val)
		if err == nil {
			redisCli.rdb.Set(ctx, catKey, cnt + 1, 0)
		}
	}
	
	createdAt, err := utils.ConvertToUnix(data["created_at"])
	if err != nil {
		return err
	}
	data["created_at"] = createdAt
	key := fmt.Sprintf("product:%s", ID)
	if err := redisCli.rdb.HSet(ctx, key, data); err.Err() != redis.Nil {
		return err.Err()
	}
	return nil
}

func (redisCli *RedisClient) GetProducts(ctx context.Context, queries string, page int) ([]FetchProduct, error) {
	pageSize := 10
	offset := pageSize * page
	result, err := redisCli.rdb.Do(ctx, "FT.SEARCH", "idx:products", queries, "SORTBY", "created_at", "ASC", "LIMIT", offset, pageSize).Result()
	if err != nil {
		return []FetchProduct{}, err
	}
	resultArray, ok := result.([]interface{})
	if !ok || len(resultArray) < 2 {
		return []FetchProduct{}, fmt.Errorf("not expected data structure")
	}
	var cacheResult []FetchProduct
	for i := 1; i < len(resultArray); i++ {
		doc, ok := resultArray[i].([]interface{})
		if !ok {
			continue
		}
		var product FetchProduct
		for j := 0; j < len(doc); j += 2 {
			key, isString := doc[j].(string)
			val := doc[j+1].(string)
			if !isString {
				continue
			}
			switch key {
			case "name":
				product.Name = val
			case "slug":
				product.Slug = val
			case "image_url":
				product.ImageUrl = val
			case "category":
				product.Category = val
			case "description":
				product.Description = val
			case "price":
				price, err := strconv.ParseFloat(val, 64)
				if err != nil {
					log.Printf("error: %s", err)
					break
				}
				product.Price = price
			case "stock_quantity":
				cnt, err := strconv.Atoi(val)
				if err != nil {
					log.Printf("error: %s", err)
					break
				}
				product.StockQuantity = cnt
			case "created_at":
				createdAt, err := utils.ConvertUnixToTime(val)
				if err != nil {
					log.Printf("error: %s", err)
					continue
				}
				product.CreatedAt = createdAt
			}
		}
		cacheResult = append(cacheResult, product)
	}
	return cacheResult, nil
}

func (redisCli *RedisClient) Listen(ctx context.Context) error {

	ticker := time.NewTicker(30 * time.Minute)

	defer func(){
		ticker.Stop()
		close(redisCli.Add)
		close(redisCli.Update)
		close(redisCli.Pop)
	}()

loop:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := redisCli.rdb.Ping(context.Background()).Err(); err != nil {
				log.Printf("Error:  %v", err)
				break loop
			}
		case payload := <-redisCli.Add:
			if err := redisCli.AddProduct(ctx, payload); err != nil {
				log.Println("err", err.Error())
			}
		case payload := <-redisCli.Update:
			ID, err := uuid.Parse(payload["id"].(string))
			if err != nil{
				log.Println("err", err.Error())
			}
			if err := redisCli.UpdateProduct(ctx, ID, payload); err != nil {
				log.Println("err: ", err.Error())
			}
			
		case payload := <-redisCli.Pop:
			ID, err := uuid.Parse(payload["id"].(string))
			if err != nil{
				log.Println("err", err.Error())
			}
			if err := redisCli.PopProduct(ctx,ID); err != nil {
				log.Println("err: ", err.Error())
			}
		}
		
	}
	panic("something wrong")
}
