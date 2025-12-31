package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	pb "github.com/subganapathy/resilient-grpc-client/e2e/proto"
)

type echoServer struct {
	pb.UnimplementedEchoServiceServer
}

func (s *echoServer) Echo(ctx context.Context, req *pb.EchoRequest) (*pb.EchoResponse, error) {
	// Simulate some processing time
	time.Sleep(10 * time.Millisecond)
	return &pb.EchoResponse{Message: req.Message}, nil
}

func (s *echoServer) EchoStream(stream pb.EchoService_EchoStreamServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}

		// Echo back the message
		if err := stream.Send(&pb.EchoResponse{Message: req.Message}); err != nil {
			return err
		}
	}
}

func main() {
	port := flag.Int("port", 50051, "gRPC server port")
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Simple gRPC server without TLS for testing
	s := grpc.NewServer(
		grpc.Creds(insecure.NewCredentials()),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    10 * time.Second,
			Timeout: 5 * time.Second,
		}),
	)

	pb.RegisterEchoServiceServer(s, &echoServer{})

	log.Printf("gRPC server listening on :%d", *port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
