package main

import (
	"fmt"
	"google.golang.org/grpc"
	"log"
	"minescale/pkg/api"
	"net"
	"os"
	"os/signal"
	"time"
)

func main() {

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 50051))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		panic(fmt.Sprintf("failed to listen: %v", err))
	}
	grpcServer := grpc.NewServer()
	api.NewServer(grpcServer)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}

	go func() {
		time.Sleep(15 * time.Second)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	grpcServer.GracefulStop()

}
