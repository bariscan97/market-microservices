package customer_client

import (
	"context"
	"time"

	pb "cart-service/grpc/pb/customer-rpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc"
)

type Client struct {
	conn   *grpc.ClientConn
	client pb.CustomerServiceClient
}

func NewClient(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials())) // WithInsecure: No TLS
	if err != nil {
		return nil, err
	}

	client := pb.NewCustomerServiceClient(conn)

	return &Client{
		conn:   conn,
		client: client,
	}, nil
}

func (c *Client) GetCustomer(customerID string) (*pb.CustomerRes, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req := &pb.CustomerReq{
		Id: customerID,
	}

	return c.client.GetCustomer(ctx, req)
}

func (c *Client) Close() error {
	return c.conn.Close()
}
