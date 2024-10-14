package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	hellopb "github.com/fchimpan/go-gprc/pkg/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type myServer struct {
	hellopb.UnimplementedGreetingServiceServer
}

func (s *myServer) SayHello(ctx context.Context, req *hellopb.HelloRequest) (*hellopb.HelloResponse, error) {
	name := req.GetName()
	msg := fmt.Sprintf("Hello, %s!", name)
	return &hellopb.HelloResponse{Message: msg}, nil
}

func (s *myServer) SayHelloStream(req *hellopb.HelloRequest, stream hellopb.GreetingService_SayHelloStreamServer) error {
	name := req.GetName()
	msg := fmt.Sprintf("Hello, %s!", name)
	for range 5 {
		res := &hellopb.HelloResponse{Message: msg}
		if err := stream.Send(res); err != nil {
			return err
		}
		time.Sleep(1 * time.Second)
	}
	// ent stream
	return nil
}

func (s *myServer) SayHelloBiStreams(stream hellopb.GreetingService_SayHelloBiStreamsServer) error {
	for {
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		name := req.GetName()
		msg := fmt.Sprintf("Hello, %s!", name)
		res := &hellopb.HelloResponse{Message: msg}
		if err := stream.Send(res); err != nil {
			return err
		}
	}
}

func main() {
	port := 8081
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	defer lis.Close()

	s := grpc.NewServer()

	// Register the service
	hellopb.RegisterGreetingServiceServer(s, &myServer{})
	reflection.Register(s)

	go func() {
		log.Println("Server is running on port:", port)
		s.Serve(lis)
	}()

	// if pressed ctrl+c, gracefully shutdown the server
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan

	log.Println("Shutting down the server...")
	s.GracefulStop()
}
