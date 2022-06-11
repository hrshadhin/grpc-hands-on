package main

import (
	"blog/pb"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"io"
	"log"
	"time"
)

func main() {
	addr := "localhost:50051"

	cc, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Couldn't connect on %s\nerr: %v", addr, err)
	}
	defer cc.Close()

	client := pb.NewBlogServiceClient(cc)
	createBlog(client)
	createBlog(client)
	curdBlog(client)
	listBlog(client)
}

func createBlog(c pb.BlogServiceClient) {
	// creating blog
	fmt.Println("Creating blog")
	timestamp := time.Now().Format(time.RFC822)
	blog := &pb.Blog{
		AuthorId: "HRS",
		Title:    "My Blog on " + timestamp,
		Content:  "Content of the my blog on " + timestamp,
	}
	res, err := c.CreateBlog(context.Background(), &pb.CreateBlogRequest{Blog: blog})
	if err != nil {
		fmt.Printf("error while creating blog: %v", err)
		return
	}

	fmt.Printf("Blog has been created %v\n", res.GetBlog())
}

func curdBlog(c pb.BlogServiceClient) {
	// creating blog
	fmt.Println("Creating blog")
	blog := &pb.Blog{
		AuthorId: "HRS",
		Title:    "My First Blog",
		Content:  "Content of the first blog",
	}
	res, err := c.CreateBlog(context.Background(), &pb.CreateBlogRequest{Blog: blog})
	if err != nil {
		fmt.Printf("error while creating blog: %v", err)
	}

	fmt.Printf("Blog has been created %v\n", res.GetBlog())

	// reading blog
	fmt.Println("Reading blog")
	blogID := res.GetBlog().GetId()
	resRes, readErr := c.ReadBlog(context.Background(), &pb.ReadBlogRequest{BlogId: blogID})
	if readErr != nil {
		fmt.Printf("error while reading blog: %v", readErr)
	}
	fmt.Printf("Blog: %v\n", resRes.GetBlog())
	_, readErr = c.ReadBlog(context.Background(), &pb.ReadBlogRequest{BlogId: "62a4f1c5974d7594be6c695a"})
	if readErr != nil {
		fmt.Printf("error while reading blog: %v\n", readErr)
	}

	// update blog
	fmt.Println("Updating blog")
	newBlog := &pb.Blog{
		Id:       blogID,
		AuthorId: "Changed Author",
		Title:    "My 3rd Blog (edited)",
		Content:  "Content of the 3rd blog",
	}
	updateRes, err := c.UpdateBlog(context.Background(), &pb.UpdateBlogRequest{Blog: newBlog})
	if err != nil {
		fmt.Printf("error while updating blog: %v", err)
	}

	fmt.Printf("Updated Blog: %v\n", updateRes.GetBlog())

	// delete blog
	deleteRes, err := c.DeleteBlog(context.Background(), &pb.DeleteBlogRequest{BlogId: blogID})
	if err != nil {
		fmt.Printf("error while deleting blog: %v", err)
	}

	fmt.Printf("Deleted Blog: %v\n", deleteRes.GetDeleted())
}

func listBlog(c pb.BlogServiceClient) {
	// creating blog
	fmt.Println("\n\nListing blog")

	stream, err := c.ListBlog(context.Background(), &pb.ListBlogRequest{})
	if err != nil {
		fmt.Printf("error while listing blog: %v", err)
		return
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("error while reading blog: %v", err)
			break
		}

		fmt.Println(res.GetBlog())
	}

}
