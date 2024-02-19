package club

import (
	"context"
	"errors"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/services/accessControl"
	clubv1 "github.com/ARUMANDESU/uniclubs-protos/gen/go/club"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MembershipService interface {
	CreateJoinRequest(ctx context.Context, userID, clubID int64) error
	ApproveMembership(ctx context.Context, clubID, userID int64) error
	RejectMembership(ctx context.Context, clubID, userID int64) error
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
	err := validation.ValidateStruct(req,
		validation.Field(&req.ClubId, validation.Required, validation.Min(1)),
		validation.Field(&req.UserId, validation.Required, validation.Min(1)),
		validation.Field(&req.MemberId, validation.Required, validation.Min(1)),
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	isAuthorized, err := s.permission.CanHandleMembershipRequest(ctx, req.GetClubId(), req.GetMemberId())
	if err != nil {
		if errors.Is(err, accessControl.ErrUserNotClubMember) {
			return nil, status.Error(codes.PermissionDenied, ErrUserNotClubMember.Error())
		}
		return nil, status.Error(codes.Internal, ErrInternal.Error())
	}
	if !isAuthorized {
		return nil, status.Error(codes.PermissionDenied, ErrUserNonAuthorized.Error())
	}

	switch req.GetAction() {
	case clubv1.HandleClubAction_APPROVE:
		err = s.membership.ApproveMembership(ctx, req.GetClubId(), req.GetUserId())
	default:
		err = s.membership.RejectMembership(ctx, req.GetClubId(), req.GetUserId())
	}
	if err != nil {
		return nil, status.Error(codes.Internal, ErrInternal.Error())
	}

	return &empty.Empty{}, nil
}

func (s serverApi) LeaveClub(context.Context, *clubv1.LeaveClubRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LeaveClub not implemented")
}
