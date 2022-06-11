package main

import (
	"blog/pb"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var collection *mongo.Collection

type server struct{}

type blogItem struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	AuthorID string             `bson:"author_id"`
	Title    string             `bson:"title"`
	Content  string             `bson:"content"`
}

func main() {
	// config variables
	dbURI := "mongodb://root:toor@localhost:27017"
	addr := "0.0.0.0:50051"

	// if we crash the go code, we get the file name & line number
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// connect to database
	log.Println("Connecting to MongoDB")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dbURI))
	if err != nil {
		log.Fatalf("Failed to connect MongoDB\nerr: %v", err)
	}
	collection = client.Database("grpc").Collection("blog")

	log.Printf("Starting the listener on %s", addr)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on %s\nerr: %v", addr, err)
	}

	opts := []grpc.ServerOption{}
	s := grpc.NewServer(opts...)
	pb.RegisterBlogServiceServer(s, &server{})

	reflection.Register(s)

	go func() {
		log.Println("Starting gRPC server...")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to start gRPC server on %s\nerr: %v", addr, err)
		}
	}()

	// signal channel to capture system calls
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	// block until a signal is received
	<-stopCh
	log.Println("Stopping gRPC server...")
	s.Stop()
	log.Println("Closing the listener")
	_ = lis.Close()
	log.Println("Closing MongoDB connection")
	_ = client.Disconnect(ctx)
	log.Println("STOPPED")
}

func (*server) CreateBlog(ctx context.Context, req *pb.CreateBlogRequest) (*pb.CreateBlogResponse, error) {
	log.Println("Create blog request")
	blog := req.GetBlog()
	data := blogItem{
		AuthorID: blog.GetAuthorId(),
		Title:    blog.GetTitle(),
		Content:  blog.GetContent(),
	}

	res, err := collection.InsertOne(context.Background(), data)
	if err != nil {
		log.Printf("Internal error: %v", err)
		return nil, status.Errorf(
			codes.Internal,
			"Internal error",
		)
	}

	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		log.Println("Cannot convert to OID")
		return nil, status.Errorf(
			codes.Internal,
			"Can not convert to OID",
		)
	}

	return &pb.CreateBlogResponse{
		Blog: &pb.Blog{
			Id:       oid.Hex(),
			AuthorId: blog.GetAuthorId(),
			Title:    blog.GetTitle(),
			Content:  blog.GetContent(),
		},
	}, nil
}

func (*server) ReadBlog(ctx context.Context, req *pb.ReadBlogRequest) (*pb.ReadBlogResponse, error) {
	log.Println("Read blog request")

	blogID := req.GetBlogId()
	oid, err := primitive.ObjectIDFromHex(blogID)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"Can not parse ID",
		)
	}

	// create an empty struct
	data := &blogItem{}
	filter := bson.D{{"_id", oid}}

	err = collection.FindOne(context.Background(), filter).Decode(&data)
	if err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			"Can not find blog with specified ID",
		)
	}

	return &pb.ReadBlogResponse{
		Blog: dataToBlogPb(data),
	}, nil

}

func dataToBlogPb(data *blogItem) *pb.Blog {
	return &pb.Blog{
		Id:       data.ID.Hex(),
		AuthorId: data.AuthorID,
		Title:    data.Title,
		Content:  data.Content,
	}
}

func (*server) UpdateBlog(ctx context.Context, req *pb.UpdateBlogRequest) (*pb.UpdateBlogResponse, error) {
	log.Println("Update blog request")

	blog := req.GetBlog()
	oid, err := primitive.ObjectIDFromHex(blog.GetId())
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"Can not parse ID",
		)
	}

	// create an empty struct
	data := &blogItem{}
	filter := bson.D{{"_id", oid}}

	err = collection.FindOne(context.Background(), filter).Decode(&data)
	if err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			"Can not find blog with specified ID",
		)
	}

	// update internal struct
	data.AuthorID = blog.GetAuthorId()
	data.Title = blog.GetTitle()
	data.Content = blog.GetContent()

	_, updateErr := collection.ReplaceOne(context.Background(), filter, data)
	if updateErr != nil {
		log.Printf("Internal error: %v", err)
		return nil, status.Errorf(
			codes.Internal, "Can not update document in mongodb",
		)
	}

	return &pb.UpdateBlogResponse{
		Blog: dataToBlogPb(data),
	}, nil
}

func (*server) DeleteBlog(ctx context.Context, req *pb.DeleteBlogRequest) (*pb.DeleteBlogResponse, error) {
	log.Println("Update blog request")

	blogID := req.GetBlogId()
	oid, err := primitive.ObjectIDFromHex(blogID)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"Can not parse ID",
		)
	}

	// create an empty struct
	data := &blogItem{}
	filter := bson.D{{"_id", oid}}
	err = collection.FindOne(context.Background(), filter).Decode(&data)
	if err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			"Can not find blog with specified ID",
		)
	}

	deleteRes, deleteErr := collection.DeleteOne(context.Background(), filter)
	if deleteErr != nil {
		log.Printf("Internal error: %v", err)
		return nil, status.Errorf(
			codes.Internal, "Can not delete document in mongodb",
		)
	}

	return &pb.DeleteBlogResponse{
		Deleted: deleteRes.DeletedCount != 0,
	}, nil
}

func (*server) ListBlog(req *pb.ListBlogRequest, stream pb.BlogService_ListBlogServer) error {
	log.Println("List blog request")

	cur, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		log.Printf("Internal error: %v", err)
		return status.Errorf(
			codes.Internal, "Can not fetch blogs from mongodb",
		)
	}
	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		data := &blogItem{}
		err := cur.Decode(data)
		if err != nil {
			log.Printf("Error while decoding error: %v", err)
			return status.Errorf(
				codes.Internal, "Error while decoding data from mongodb",
			)
		}

		_ = stream.Send(&pb.ListBlogResponse{Blog: dataToBlogPb(data)})

	}

	if err := cur.Err(); err != nil {
		log.Printf("Unkown internal error: %v", err)
		return status.Errorf(
			codes.Internal, "Unknown internal error",
		)
	}

	return nil
}
