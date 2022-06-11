package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"greet/pb"
	"io"
	"log"
	"net"
	"strconv"
	"time"

	"google.golang.org/grpc"
)

type server struct{}

func main() {
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{}

	tls := true
	if tls {
		certFile := "ssl/server.crt"
		keyFile := "ssl/server.pem"
		creds, sslErr := credentials.NewServerTLSFromFile(certFile, keyFile)
		if sslErr != nil {
			log.Fatalf("failed loading certificates: %v", sslErr)
		}

		opts = append(opts, grpc.Creds(creds))
	}

	s := grpc.NewServer(opts...)
	pb.RegisterGreetServiceServer(s, &server{})

	fmt.Println("starting gRPC server...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (*server) Greet(ctx context.Context, req *pb.GreetRequest) (*pb.GreetResponse, error) {
	fmt.Printf("Greet function was invoked with %v\n", req)
	firstName := req.GetGreeting().GetFirstName()
	//lastName := req.GetGreeting().GetLastName()

	result := "Hello " + firstName
	res := &pb.GreetResponse{
		Result: result,
	}

	return res, nil
}

func (*server) GreetManyTimes(req *pb.GreetManyTimesRequest, stream pb.GreetService_GreetManyTimesServer) error {
	fmt.Printf("Greet many times function was invoked with %v\n", req)
	firstName := req.GetGreeting().GetFirstName()
	for i := 1; i <= 10; i++ {
		res := &pb.GreetManyTimesResponse{
			Result: "Hello " + firstName + " number " + strconv.Itoa(i),
		}
		err := stream.Send(res)
		if err != nil {
			log.Fatalf("couldn't send stream %v", err)
		}
		time.Sleep(1000 * time.Millisecond)
	}

	return nil
}

func (*server) LongGreet(stream pb.GreetService_LongGreetServer) error {
	fmt.Println("LongGreet function was invoked with a streaming request")
	result := ""

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			// we have finished reading the client stream
			return stream.SendAndClose(&pb.LongGreetResponse{
				Result: result,
			})
		}

		if err != nil {
			log.Fatalf("error while reading client stream: %v", err)
		}

		firstName := req.GetGreeting().GetFirstName()
		result += "Hello " + firstName + "! "
	}
}

func (*server) GreetEveryOne(stream pb.GreetService_GreetEveryOneServer) error {
	fmt.Println("GreetEveryOne function was invoked with a streaming request")

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("error while reading client stream: %v", err)
		}
		firstName := req.GetGreeting().GetFirstName()
		result := "Hello " + firstName + "! "
		err = stream.Send(&pb.GreetEveryoneResponse{
			Result: result,
		})

		if err != nil {
			log.Fatalf("error while sending data to client: %v", err)
		}

	}
}

func (*server) GreetWithDeadline(ctx context.Context, req *pb.GreetWithDeadlineRequest) (*pb.GreetWithDeadlineResponse, error) {
	fmt.Println("GreetWithDeadline function was invoked with deadline")

	for i := 0; i < 3; i++ {
		if ctx.Err() == context.Canceled {
			fmt.Println("The client canceled the request!")
			return nil, status.Error(codes.Canceled, "the client canceled the request")
		}
		time.Sleep(1 * time.Second)
	}

	firstName := req.GetGreeting().GetFirstName()
	result := "Hello " + firstName + "! "
	res := &pb.GreetWithDeadlineResponse{
		Result: result,
	}

	return res, nil
}
