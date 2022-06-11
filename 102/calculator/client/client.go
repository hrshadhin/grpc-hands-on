package main

import (
	"calculator/pb"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
)

func main() {
	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could connect: %v", err)
	}
	defer cc.Close()

	client := pb.NewCalculatorServiceClient(cc)
	//doUnary(client)
	doErrorUnary(client)
}

func doUnary(c pb.CalculatorServiceClient) {
	req := &pb.SumRequest{
		FirstNumber:  5,
		SecondNumber: 3,
	}
	res, err := c.Sum(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling calculator RPC: %v", err)
	}

	fmt.Printf("Sum of %d & %d is = %d", req.FirstNumber, req.SecondNumber, res.SumResult)
}

func doErrorUnary(c pb.CalculatorServiceClient) {
	fmt.Println("starting  to do a SquareRoot Unary RPC...")

	// correct call
	doErrorCall(c, 25)

	// error call
	doErrorCall(c, -21)
}

func doErrorCall(c pb.CalculatorServiceClient, n int32) {
	req := &pb.SquareRootRequest{
		Number: n,
	}

	res, err := c.SquareRoot(context.Background(), req)
	if err != nil {
		respErr, ok := status.FromError(err)
		if ok {
			// actual error from gRPC (user error)
			fmt.Printf("error message from server: %v\n", respErr.Message())
			fmt.Printf("error code from server: %v\n", respErr.Code())
			if respErr.Code() == codes.InvalidArgument {
				fmt.Println("we probably sent a negative number!")
				return
			}
		} else {
			log.Fatalf("Big error calling SquareRoot: %v", err)
			return
		}
	}

	fmt.Printf("SquareRoot of %d is = %.2f\n-0-\n", req.GetNumber(), res.GetNumberRoot())
}
