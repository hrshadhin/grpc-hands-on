package main

import (
	"calculator/pb"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"log"
	"math"
	"net"
)

type server struct{}

func (*server) Sum(ctx context.Context, req *pb.SumRequest) (*pb.SumResponse, error) {
	fmt.Println("invoked sum function")
	res := &pb.SumResponse{
		SumResult: req.FirstNumber + req.SecondNumber,
	}

	return res, nil
}

func (*server) SquareRoot(ctx context.Context, req *pb.SquareRootRequest) (*pb.SquareRootResponse, error) {
	fmt.Println("invoked SquareRoot function")
	number := req.GetNumber()
	if number < 0 {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Received a negative number: %v", number),
		)
	}

	return &pb.SquareRootResponse{
		NumberRoot: math.Sqrt(float64(number)),
	}, nil

}

func main() {
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("failed to listen err: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterCalculatorServiceServer(s, &server{})

	// Register reflection service on gRPC server.
	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve err: %v", err)
	}
}
