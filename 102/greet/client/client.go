package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"greet/pb"
	"io"
	"log"
	"time"
)

func main() {
	opts := grpc.WithInsecure()

	tls := true
	if tls {
		certFile := "ssl/ca.crt" // Certificate Authority Trust certificate
		creds, sslErr := credentials.NewClientTLSFromFile(certFile, "")
		if sslErr != nil {
			log.Fatalf("error while loading CA trust certificate: %v", sslErr)
		}

		opts = grpc.WithTransportCredentials(creds)
	}

	cc, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	defer cc.Close()

	client := pb.NewGreetServiceClient(cc)
	doUnary(client)
	doServerStream(client)
	doClientStreaming(client)
	doBiDiStreaming(client)
	doUnaryWithDeadline(client, 1*time.Second) // should complete
	doUnaryWithDeadline(client, 5*time.Second) // should time out

}

func doUnary(c pb.GreetServiceClient) {
	fmt.Println("starting to do a Unary RPC...")
	req := &pb.GreetRequest{
		Greeting: &pb.Greeting{
			FirstName: "HR",
			LastName:  "Shadhin",
		},
	}
	res, err := c.Greet(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling Greet RPC: %v", err)
	}

	fmt.Printf("Response from Greet: %v", res.Result)
}

func doServerStream(c pb.GreetServiceClient) {
	fmt.Println("starting to do a server stream RPC...")
	req := &pb.GreetManyTimesRequest{
		Greeting: &pb.Greeting{
			FirstName: "HR",
			LastName:  "Shadhin",
		},
	}
	resStream, err := c.GreetManyTimes(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling server streaming Greet many time RPC: %v", err)
	}

	for {
		msg, err := resStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("error while reading stream: %v", err)

		}

		fmt.Printf("Response from Greet many time: %s\n", msg.Result)

	}

}

func doClientStreaming(c pb.GreetServiceClient) {
	fmt.Println("starting to do a Client streaming RPC...")

	requests := []*pb.LongGreetRequest{
		&pb.LongGreetRequest{
			Greeting: &pb.Greeting{
				FirstName: "HR",
				LastName:  "Shadhin",
			},
		},
		&pb.LongGreetRequest{
			Greeting: &pb.Greeting{
				FirstName: "Foo",
				LastName:  "Bar",
			},
		},
		&pb.LongGreetRequest{
			Greeting: &pb.Greeting{
				FirstName: "Piper",
			},
		},
	}

	stream, err := c.LongGreet(context.Background())
	if err != nil {
		log.Fatalf("error while calling longGreet: %v", err)
	}

	for _, req := range requests {
		fmt.Printf("sending req: %v\n", req)
		_ = stream.Send(req)
		time.Sleep(1 * time.Second)
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("error while receiving response from LongGreet: %v", err)
	}

	fmt.Printf("Response: %s\n", res.GetResult())
}

func doBiDiStreaming(c pb.GreetServiceClient) {
	fmt.Println("starting to do a BiDi streaming RPC...")

	// we create a stream by invoking the client
	stream, err := c.GreetEveryOne(context.Background())
	if err != nil {
		log.Fatalf("error while creating stream: %v", err)
	}

	requests := []*pb.GreetEveryoneRequest{
		&pb.GreetEveryoneRequest{
			Greeting: &pb.Greeting{
				FirstName: "HR",
				LastName:  "Shadhin",
			},
		},
		&pb.GreetEveryoneRequest{
			Greeting: &pb.Greeting{
				FirstName: "Foo",
				LastName:  "Bar",
			},
		},
		&pb.GreetEveryoneRequest{
			Greeting: &pb.Greeting{
				FirstName: "Piper",
			},
		},
	}
	waitc := make(chan struct{})

	// we send a bunch of messages to server (go routine)
	go func() {
		for _, req := range requests {
			fmt.Printf("sening message: %v\n", req)
			err := stream.Send(req)
			if err != nil {
				log.Fatalf("error while sending stream request: %v", err)
			}
			time.Sleep(1 * time.Second)
		}

		err := stream.CloseSend()
		if err != nil {
			log.Fatalf("error while closing stream: %v", err)
		}
	}()

	// we receive a bunch of messages from the server (go routine)
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("error while receiving stream: %v", err)
				break
			}
			fmt.Printf("Received: %v\n", res.GetResult())
		}

		close(waitc)

	}()

	// block until everything is done
	<-waitc
}

func doUnaryWithDeadline(c pb.GreetServiceClient, timeout time.Duration) {
	fmt.Println("starting to do a Unary GreetWithDeadline RPC...")
	req := &pb.GreetWithDeadlineRequest{
		Greeting: &pb.Greeting{
			FirstName: "HR",
			LastName:  "Shadhin",
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res, err := c.GreetWithDeadline(ctx, req)
	if err != nil {
		statusErr, ok := status.FromError(err)
		if ok {
			if statusErr.Code() == codes.DeadlineExceeded {
				fmt.Println("Timeout was hit! Deadline was exceeded")
			} else {
				fmt.Printf("uunexpected error: %v", statusErr)
			}
		} else {
			log.Fatalf("error while calling GreetWithDeadline RPC: %v", err)
		}
		return
	}

	fmt.Printf("Response: %v", res.GetResult())
}
