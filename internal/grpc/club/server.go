package club

import (
	"errors"
	clubv1 "github.com/ARUMANDESU/uniclubs-protos/gen/go/club"
	"google.golang.org/grpc"
)

var (
	ErrInternal     = errors.New("internal error")
	ErrClubNotFound = errors.New("club not found")
)

type serverApi struct {
	clubv1.UnimplementedClubServer
	management ManagementService
	info       InfoService
	membership MembershipService
}

func Register(
	gRPC *grpc.Server,
	management ManagementService,
	membership MembershipService,
	info InfoService,
) {
	clubv1.RegisterClubServer(gRPC, &serverApi{
		management: management,
		membership: membership,
		info:       info,
	})
}
