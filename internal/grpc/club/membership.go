package club

import (
	"context"
	clubv1 "github.com/ARUMANDESU/uniclubs-protos/gen/go/club"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MembershipService interface {
	CreateJoinRequest(ctx context.Context, userID, clubID int64) error
	ApproveMembership(ctx context.Context, userID, memberID, clubID int64) error
	RejectMembership(ctx context.Context, userID, memberID, clubID int64) error
}

func (s serverApi) RequestToJoinClub(ctx context.Context, req *clubv1.RequestToJoinClubRequest) (*empty.Empty, error) {
	err := validation.ValidateStruct(req,
		validation.Field(&req.ClubId, validation.Required, validation.Min(1)),
		validation.Field(&req.UserId, validation.Required, validation.Min(1)),
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = s.membership.CreateJoinRequest(ctx, req.GetUserId(), req.GetClubId())
	if err != nil {
		return nil, status.Error(codes.Internal, ErrInternal.Error())
	}

	return &empty.Empty{}, nil
}

func (s serverApi) HandleJoinClub(ctx context.Context, req *clubv1.HandleJoinClubRequest) (*empty.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (s serverApi) LeaveClub(context.Context, *clubv1.LeaveClubRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LeaveClub not implemented")
}
