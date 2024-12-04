package product_client

import (
	"context"
	"time"

	pb "cart-service/grpc/pb/product-rpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc"
)

type Client struct {
	conn   *grpc.ClientConn
	client pb.ProductServiceClient
}

func NewClient(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials())) 
	if err != nil {
		return nil, err
	}

	client := pb.NewProductServiceClient(conn)

	return &Client{
		conn:   conn,
		client: client,
	}, nil
}

func (c *Client) GetCustomer(ids []string) (*pb.ProductRes, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req := &pb.ProductReq{
		Id: ids,
	}

	return c.client.AllItemsExists(ctx, req)
}

func (c *Client) Close() error {
	return c.conn.Close()
}
