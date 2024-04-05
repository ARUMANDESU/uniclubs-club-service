package club

import (
	"context"
	"errors"
	clubv1 "github.com/ARUMANDESU/uniclubs-protos/gen/go/club"
	"google.golang.org/grpc"
)

var (
	ErrInternal            = errors.New("internal error")
	ErrClubNotFound        = errors.New("club not found")
	ErrUserNotClubMember   = errors.New("user is not club member or club_id is not correct")
	ErrTargetNotClubMember = errors.New("target user is not club member")
	ErrUserNonAuthorized   = errors.New("user does not have permission")
	ErrEditConflict        = errors.New("unable to update the record due to an edit conflict, please try again")
)

type serverApi struct {
	clubv1.UnimplementedClubServer
	management ManagementService
	info       InfoService
	membership MembershipService
	permission PermissionService
}

type PermissionService interface {
	CanActOnMember(ctx context.Context, clubID, userID, targetID int64, permission uint64) (bool, error)
	CanHandleMembershipRequest(ctx context.Context, clubID, userID int64) (bool, error)
	CanManageClub(ctx context.Context, clubID, userID int64) (bool, error)
}

func Register(
	gRPC *grpc.Server,
	management ManagementService,
	membership MembershipService,
	info InfoService,
	permission PermissionService,
) {
	clubv1.RegisterClubServer(gRPC, &serverApi{
		management: management,
		membership: membership,
		info:       info,
		permission: permission,
	})
}
