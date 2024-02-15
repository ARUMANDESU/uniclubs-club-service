package club

import (
	"context"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/domain/dtos"
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

func (s serverApi) DeactivateClub(ctx context.Context, req *clubv1.DeactivateClubRequest) (*empty.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (s serverApi) UpdateClub(ctx context.Context, req *clubv1.UpdateClubRequest) (*empty.Empty, error) {
	//TODO implement me
	panic("implement me")
}
