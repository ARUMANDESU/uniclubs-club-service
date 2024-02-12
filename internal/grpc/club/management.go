package club

import (
	"context"
	"errors"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/domain"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/domain/dtos"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/services/management"
	clubv1 "github.com/ARUMANDESU/uniclubs-protos/gen/go/club"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ManagementService interface {
	CreateClub(ctx context.Context, dto dtos.CreateClubDTO) error
	ApproveClub(ctx context.Context, clubID int64) error
	RejectClub(ctx context.Context, clubID int64) error
	GetClub(ctx context.Context, clubID int64) (*domain.Club, error)
	ListClub(
		ctx context.Context,
		query string, clubTypes []string,
		filters domain.Filters,
	) (
		[]*domain.Club,
		*domain.Metadata,
		error,
	)
	ListNotActivatedClubs(
		ctx context.Context,
		query string, clubTypes []string,
		filters domain.Filters,
	) (
		[]*domain.ClubUser,
		*domain.Metadata,
		error,
	)
	CreateJoinRequest(ctx context.Context, userID, clubID int64) error
	ApproveMembership(ctx context.Context, userID, memberID, clubID int64) error
	RejectMembership(ctx context.Context, userID, memberID, clubID int64) error
}

func (s serverApi) CreateClub(ctx context.Context, req *clubv1.CreateClubRequest) (*empty.Empty, error) {
	err := validation.ValidateStruct(req,
		validation.Field(&req.OwnerId, validation.Required, validation.Min(1)),
		validation.Field(&req.Name, validation.Required, validation.Length(3, 250)),
		validation.Field(&req.ClubType, validation.Required, validation.Length(3, 250)),
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = s.management.CreateClub(ctx, dtos.CreateClubRequestToDTO(req))
	if err != nil {
		return nil, status.Error(codes.Internal, ErrInternal.Error())
	}

	return &empty.Empty{}, nil

}

func (s serverApi) HandleNewClub(ctx context.Context, req *clubv1.HandleNewClubRequest) (*empty.Empty, error) {
	err := validation.ValidateStruct(req,
		validation.Field(&req.ClubId, validation.Required, validation.Min(1)),
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if req.GetAction() == clubv1.HandleClubAction_APPROVE {
		err = s.management.ApproveClub(ctx, req.GetClubId())
	} else {
		err = s.management.RejectClub(ctx, req.GetClubId())
	}
	if err != nil {
		return nil, status.Error(codes.Internal, ErrInternal.Error())
	}

	return &empty.Empty{}, nil
}

func (s serverApi) GetClub(ctx context.Context, req *clubv1.GetClubRequest) (*clubv1.ClubObject, error) {
	err := validation.Validate(req.ClubId, validation.Required, validation.Min(1))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	club, err := s.management.GetClub(ctx, req.GetClubId())
	if err != nil {
		if errors.Is(err, management.ErrClubNotExists) {
			return nil, status.Error(codes.NotFound, ErrClubNotFound.Error())
		}
		return nil, status.Error(codes.Internal, ErrInternal.Error())
	}

	return domain.ClubToClubObject(club), nil
}

func (s serverApi) ListClubs(ctx context.Context, req *clubv1.ListClubRequest) (*clubv1.ListClubResponse, error) {
	err := validation.ValidateStruct(req,
		validation.Field(&req.PageNumber, validation.Required, validation.Min(1)),
		validation.Field(&req.PageSize, validation.Required, validation.Min(1)),
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	f := domain.Filters{
		Page:     req.GetPageNumber(),
		PageSize: req.GetPageSize(),
	}

	clubs, metadata, err := s.management.ListClub(ctx, req.GetQuery(), req.GetClubType(), f)
	if err != nil {
		return nil, status.Error(codes.Internal, ErrInternal.Error())
	}

	return &clubv1.ListClubResponse{
		Clubs:    domain.MapClubArrToClubObjectArr(clubs),
		Metadata: domain.ToPagination(metadata),
	}, nil

}

func (s serverApi) ListNotActivatedClubs(ctx context.Context, req *clubv1.ListNotActivatedClubsRequest) (*clubv1.ListNotActivatedClubsResponse, error) {
	err := validation.ValidateStruct(req,
		validation.Field(&req.PageNumber, validation.Required, validation.Min(1)),
		validation.Field(&req.PageSize, validation.Required, validation.Min(1)),
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	f := domain.Filters{
		Page:     req.GetPageNumber(),
		PageSize: req.GetPageSize(),
	}

	clubsUsers, metadata, err := s.management.ListNotActivatedClubs(ctx, req.GetQuery(), req.GetClubType(), f)
	if err != nil {
		return nil, status.Error(codes.Internal, ErrInternal.Error())
	}

	return &clubv1.ListNotActivatedClubsResponse{
		List:     domain.MapClubUserArrToClubList(clubsUsers),
		Metadata: domain.ToPagination(metadata),
	}, nil

}

func (s serverApi) RequestToJoinClub(ctx context.Context, req *clubv1.RequestToJoinClubRequest) (*empty.Empty, error) {
	err := validation.ValidateStruct(req,
		validation.Field(&req.ClubId, validation.Required, validation.Min(1)),
		validation.Field(&req.UserId, validation.Required, validation.Min(1)),
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = s.management.CreateJoinRequest(ctx, req.GetUserId(), req.GetClubId())
	if err != nil {
		return nil, status.Error(codes.Internal, ErrInternal.Error())
	}

	return &empty.Empty{}, nil
}

func (s serverApi) HandleJoinClub(ctx context.Context, req *clubv1.HandleJoinClubRequest) (*empty.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (s serverApi) DeactivateClub(ctx context.Context, req *clubv1.DeactivateClubRequest) (*empty.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (s serverApi) UpdateClub(ctx context.Context, req *clubv1.UpdateClubRequest) (*empty.Empty, error) {
	//TODO implement me
	panic("implement me")
}
