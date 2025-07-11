package grpc

import (
	"context"

	pvz_v1 "github.com/PaulLocust/protos/gen/go/pvz"
	"google.golang.org/grpc"
)

type serverAPI struct {
	pvz_v1.UnimplementedPVZServiceServer
}

func Register(gRPC *grpc.Server) {
	pvz_v1.RegisterPVZServiceServer(gRPC, &serverAPI{})
}

func (s *serverAPI) GetPVZList(ctx context.Context, req *pvz_v1.GetPVZListRequest) (*pvz_v1.GetPVZListResponse, error) {
	panic("Implement me")
}
