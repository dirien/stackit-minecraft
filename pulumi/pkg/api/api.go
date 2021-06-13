package api

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"minescale/internal/grpc/minescale/minescale"
	"minescale/pkg/automation"
	"time"
)

type Server struct {
	grpcServer *grpc.Server
	automation automation.Automation
	minescale.UnimplementedMinescaleServerServer
}

func NewServer(grpcServer *grpc.Server) *Server {
	a := automation.NewAutomationProgram()
	server := &Server{
		grpcServer: grpcServer,
		automation: a,
	}
	minescale.RegisterMinescaleServerServer(grpcServer, server)
	return server
}

func (s *Server) ListMinescaleServer(*emptypb.Empty, minescale.MinescaleServer_ListMinescaleServerServer) error {
	/*log.Debugf("received server stream command")
	for i := 0; i < 5; i++ {
		err := stream.Send(&rpc.Message{
			Message: fmt.Sprintf("Sending stream back to client, iteration: %d", i),
		})
		if err != nil {
			return err
		}

		time.Sleep(2 * time.Second)
	}

	return nil*/
}

func (s *Server) CreateMinescaleServer(ctx context.Context, request *minescale.MinescaleRequest) (*minescale.MinescaleReply, error) {
	log.Printf("Received: %v", request)
	//name, flavor, cidr, pubkey
	vm, err := s.automation.CreateServer(request.GetName(), request.GetFlavor(), request.GetCidr(), request.GetPubkey())

	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &minescale.MinescaleReply{Ip: vm, Status: true}, nil
}
