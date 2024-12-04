package handler

import (
	"context"
	"net"
	"log"
	"google.golang.org/grpc"
	pb "customer-service/grpc/pb"
	"customer-service/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/google/uuid"
)

type Server struct {
	pb.UnimplementedCustomerServiceServer
	CustomerService service.ICustomerService
}

func NewServer(CustomerService service.ICustomerService) *Server {
	return &Server{
		CustomerService: CustomerService,
	}
}

func (server *Server) GetCustomer(ctx context.Context, req *pb.CustomerReq) (*pb.CustomerRes, error) {
	ID := req.Id
	parsedID, err := uuid.Parse(ID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid UUID format")
	}
	
	customer, err := server.CustomerService.GetCustomerById(parsedID)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			return nil, status.Error(codes.NotFound, "customer not found")
		}
		return nil, status.Error(codes.Internal, "internal server error")
	}
	
	if customer.FirstName == "" ||  customer.LastName == "" || customer.Address == ""  || customer.PhoneNumber == "" {
		return nil, status.Error(codes.NotFound, "not all fields are filled")
	}

	return &pb.CustomerRes{
		Firstname: customer.FirstName,
		Lastname: customer.LastName,
		Address: customer.Address,
		Phone: customer.PhoneNumber,
	}, nil
}

func (server *Server) Run(addr string) error {
	listener, err := net.Listen("tcp", addr)
	
	if err != nil {
		log.Printf("Failed to listen on %s: %v", addr, err)
		return err
	}

	grpcServer := grpc.NewServer()

	pb.RegisterCustomerServiceServer(grpcServer, server)

	log.Printf("gRPC server is running on %s", addr)
	
	return grpcServer.Serve(listener)
}