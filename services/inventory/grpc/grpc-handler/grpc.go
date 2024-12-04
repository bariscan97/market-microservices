package grpc

import (
	"context"
	"fmt"
	"inventory-service/grpc/pb"
	"inventory-service/repository"
	"log"
	"net"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GrpcServer struct {
	repo repository.IProductRepository
	pb.UnimplementedProductServiceServer
}

func NewGrpcServer(repo repository.IProductRepository) *GrpcServer{
	return &GrpcServer{
		repo: repo,
	}
}

func (server *GrpcServer) AllItemsExists(ctx context.Context, req *pb.ProductReq) (*pb.ProductRes, error) {
	for _, id := range req.Id {
		parsedID, err :=uuid.Parse(id)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid UUID format")
		}
		item ,err := server.repo.GetProductById(parsedID)
		if !item.Is_active {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("id:%s ,prodcuct not active", id)) 
		}
		if err != nil {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("id:%s ,prodcuct not found", id))
		}
	}
	return &pb.ProductRes{
		AllExist: true,
	}, nil
}

func (server *GrpcServer) Run(addr string) error {
	listener, err := net.Listen("tcp", addr)
	
	if err != nil {
		log.Printf("Failed to listen on %s: %v", addr, err)
		return err
	}

	grpcServer := grpc.NewServer()

	pb.RegisterProductServiceServer(grpcServer, server)

	log.Printf("gRPC server is running on %s", addr)
	
	return grpcServer.Serve(listener)
}