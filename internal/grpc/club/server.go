package club

import (
	"errors"
	clubv1 "github.com/ARUMANDESU/uniclubs-protos/gen/go/club"
	"google.golang.org/grpc"
)

var (
	ErrInternal = errors.New("internal error")
)

type serverApi struct {
	clubv1.UnimplementedClubServer
	management ManagementService
}

func Register(gRPC *grpc.Server, management ManagementService) {
	clubv1.RegisterClubServer(gRPC, &serverApi{management: management})
}
