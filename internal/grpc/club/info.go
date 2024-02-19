package club

import (
	"context"
	"errors"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/domain"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/services/management"
	clubv1 "github.com/ARUMANDESU/uniclubs-protos/gen/go/club"
	validation "github.com/go-ozzo/ozzo-validation"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InfoService interface {
	GetClub(ctx context.Context, clubID int64) (*domain.Club, error)
	GetUserClubs(ctx context.Context, userID int64) ([]*domain.Club, error)
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
	ListClubMembers(ctx context.Context, clubID int64, filters domain.Filters) ([]*domain.User, *domain.Metadata, error)
	ListClubJoinReq(ctx context.Context, clubID int64, filters domain.Filters) ([]*domain.User, *domain.Metadata, error)
}

func (s serverApi) GetClub(ctx context.Context, req *clubv1.GetClubRequest) (*clubv1.ClubObject, error) {
	err := validation.Validate(req.ClubId, validation.Required, validation.Min(1))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	club, err := s.info.GetClub(ctx, req.GetClubId())
	if err != nil {
		if errors.Is(err, management.ErrClubNotExists) {
			return nil, status.Error(codes.NotFound, ErrClubNotFound.Error())
		}
		return nil, status.Error(codes.Internal, ErrInternal.Error())
	}

	return club.ToClubObject(), nil
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

	clubs, metadata, err := s.info.ListClub(ctx, req.GetQuery(), req.GetClubType(), f)
	if err != nil {
		return nil, status.Error(codes.Internal, ErrInternal.Error())
	}

	return &clubv1.ListClubResponse{
		Clubs:    domain.MapClubArrToClubObjectArr(clubs),
		Metadata: domain.ToPagination(metadata),
	}, nil

}

func (s serverApi) ListNotApprovedClubs(ctx context.Context, req *clubv1.ListNotApprovedClubsRequest) (*clubv1.ListNotApprovedClubsResponse, error) {
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

	clubsUsers, metadata, err := s.info.ListNotActivatedClubs(ctx, req.GetQuery(), req.GetClubType(), f)
	if err != nil {
		return nil, status.Error(codes.Internal, ErrInternal.Error())
	}

	return &clubv1.ListNotApprovedClubsResponse{
		List:     domain.MapClubUserArrToClubList(clubsUsers),
		Metadata: domain.ToPagination(metadata),
	}, nil

}

func (s serverApi) GetUserClubs(ctx context.Context, req *clubv1.GetUserClubsRequest) (*clubv1.GetUserClubsResponse, error) {
	err := validation.Validate(req.UserId, validation.Required, validation.Min(1))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	clubs, err := s.info.GetUserClubs(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, ErrInternal.Error())
	}

	return &clubv1.GetUserClubsResponse{Clubs: domain.MapClubArrToClubObjectArr(clubs)}, nil
}

func (s serverApi) ListClubMembers(ctx context.Context, req *clubv1.ListClubMembersRequest) (*clubv1.ListClubMembersResponse, error) {
	err := validation.ValidateStruct(req,
		validation.Field(&req.ClubId, validation.Required, validation.Min(1)),
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

	members, metadata, err := s.info.ListClubMembers(ctx, req.GetClubId(), f)
	if err != nil {
		return nil, status.Error(codes.Internal, ErrInternal.Error())
	}

	return &clubv1.ListClubMembersResponse{
		Users:    domain.MapUserArrToUserObjectArr(members),
		Metadata: domain.ToPagination(metadata),
	}, nil
}

func (s serverApi) ListJoinRequests(ctx context.Context, req *clubv1.ListJoinRequestsRequest) (*clubv1.ListJoinRequestsResponse, error) {
	err := validation.ValidateStruct(req,
		validation.Field(&req.ClubId, validation.Required, validation.Min(1)),
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

	users, metadata, err := s.info.ListClubJoinReq(ctx, req.GetClubId(), f)
	if err != nil {
		return nil, status.Error(codes.Internal, ErrInternal.Error())
	}
	return &clubv1.ListJoinRequestsResponse{
		Users:    domain.MapUserArrToUserObjectArr(users),
		Metadata: domain.ToPagination(metadata),
	}, nil
}
