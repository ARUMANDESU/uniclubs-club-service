package club

import (
	"context"
	"errors"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/domain"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/domain/dtos"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/services/accessControl"
	clubv1 "github.com/ARUMANDESU/uniclubs-protos/gen/go/club"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
)

type MembershipService interface {
	CreateJoinRequest(ctx context.Context, userID, clubID int64) error
	ApproveMembership(ctx context.Context, clubID, userID int64) error
	RejectMembership(ctx context.Context, clubID, userID int64) error
	CreateNewRole(ctx context.Context, dto dtos.CreateRoleDTO) (*domain.Role, error)
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

func (s serverApi) CreateRole(ctx context.Context, req *clubv1.CreateRoleRequest) (*clubv1.Role, error) {
	err := validation.ValidateStruct(req,
		validation.Field(&req.ClubId, validation.Required, validation.Min(1)),
		validation.Field(&req.UserId, validation.Required, validation.Min(1)),
		validation.Field(&req.Name, validation.Required, validation.Length(4, 75)),
		validation.Field(&req.Permissions, validation.Each(validation.In(
			"Administrator",
			"ManageClub",
			"ManageMembership",
			"KickMember",
			"BanMember",
			"ManageRoles"))),
		validation.Field(&req.Position, validation.Required, validation.Min(1)),
		validation.Field(&req.Color, validation.Required),
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	isAuthorized, err := s.permission.CanManageRoles(ctx, req.GetClubId(), req.GetUserId())
	if err != nil {
		if errors.Is(err, accessControl.ErrUserNotClubMember) {
			return nil, status.Error(codes.PermissionDenied, ErrUserNotClubMember.Error())
		}
		return nil, status.Error(codes.Internal, ErrInternal.Error())
	}
	if !isAuthorized {
		return nil, status.Error(codes.PermissionDenied, ErrUserNonAuthorized.Error())
	}

	hexPerms, err := domain.StringArrToHex(req.Permissions)
	if err != nil {
		return nil, status.Error(codes.Internal, ErrInternal.Error())
	}

	dto := dtos.CreateRoleDTO{
		ClubID:      req.GetClubId(),
		UserID:      req.GetUserId(),
		Name:        req.GetName(),
		Position:    1,
		Permissions: strconv.FormatUint(hexPerms, 10),
		Color:       req.GetColor(),
	}

	role, err := s.membership.CreateNewRole(ctx, dto)
	if err != nil {
		return nil, status.Error(codes.Internal, ErrInternal.Error())
	}

	return role.ToRoleProto(), nil

}

func (s serverApi) UpdateRole(ctx context.Context, req *clubv1.UpdateRoleRequest) (*clubv1.Role, error) {
	//TODO implement me
	panic("implement me")
}

func (s serverApi) DeleteRole(ctx context.Context, req *clubv1.DeleteRoleRequest) (*empty.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (s serverApi) ChangeRolesPosition(ctx context.Context, request *clubv1.ChangeRolesPositionRequest) (*clubv1.ChangeRolesPositionResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s serverApi) AddRoleMembers(ctx context.Context, request *clubv1.AddRoleMembersRequest) (*clubv1.AddRoleMembersResponses, error) {
	//TODO implement me
	panic("implement me")
}
