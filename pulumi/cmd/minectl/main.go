package main

import (
	"context"
	"flag"
	"google.golang.org/grpc"
	"io/ioutil"
	"log"
	"minescale/internal/grpc/minescale/minescale"
	"time"
)

const (
	address     = "localhost:50051"
	defaultName = "world"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := minescale.NewMinescaleServerClient(conn)

	var deadlineMs = flag.Int("deadline_ms", 60*1000, "Default deadline in milliseconds.")

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*deadlineMs)*time.Millisecond)

	//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	pubKeyFile, err := ioutil.ReadFile("/Users/dirien/Tools/repos/stackit-minecraft/minecraft/ssh/minecraft.pub")

	r, err := c.CreateMinescaleServer(ctx, &minescale.MinescaleRequest{
		Name:   "automation-api",
		Cidr:   "10.2.10.0/24",
		Flavor: "c1.3",
		Pubkey: string(pubKeyFile),
	})

	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("IP: %s", r.Ip)
}
