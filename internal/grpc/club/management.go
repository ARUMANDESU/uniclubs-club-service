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
	// interface
	CreateClub(ctx context.Context, dto dtos.CreateClubDTO) error
	ApproveClub(ctx context.Context, clubID int64) error
	RejectClub(ctx context.Context, clubID int64) error
	UpdateClub(ctx context.Context, club *domain.Club) error
	UpdateLogo(ctx context.Context, clubID int64, logoUrl string) (*domain.Club, string, error)
	UpdateBanner(ctx context.Context, clubID int64, bannerUrl string) (*domain.Club, string, error)
	TransferOwnership(ctx context.Context, clubID, userID, targetID int64) error
	DeleteClub(ctx context.Context, clubID int64) error
}

func (s serverApi) DeleteClub(ctx context.Context, req *clubv1.DeleteClubRequest) (*empty.Empty, error) {
	err := validation.ValidateStruct(req,
		validation.Field(&req.ClubId, validation.Required, validation.Min(1)),
		validation.Field(&req.UserId, validation.Required, validation.Min(1)),
		validation.Field(&req.CanDelete, validation.Required),
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	isAuthorized, err := s.permission.CanDeleteClub(ctx, req.GetClubId(), req.GetUserId(), req.GetCanDelete())
	if err != nil {
		if errors.Is(err, domain.ErrUserNotClubOwner) {
			return nil, status.Error(codes.PermissionDenied, domain.ErrUserNotClubOwner.Error())
		}
		return nil, status.Error(codes.Internal, ErrInternal.Error())
	}
	if !isAuthorized {
		return nil, status.Error(codes.PermissionDenied, ErrUserNonAuthorized.Error())
	}

	err = s.management.DeleteClub(ctx, req.GetClubId())
	if err != nil {
		switch {
		case errors.Is(err, management.ErrClubNotExists):
			return nil, status.Error(codes.NotFound, ErrClubNotFound.Error())
		case errors.Is(err, domain.ErrUserNotClubOwner):
			return nil, status.Error(codes.PermissionDenied, domain.ErrUserNotClubOwner.Error())
		default:
			return nil, status.Error(codes.Internal, ErrInternal.Error())
		}
	}

	return &empty.Empty{}, nil
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
		if errors.Is(err, management.ErrClubNotExists) {
			return nil, status.Error(codes.NotFound, ErrClubNotFound.Error())
		}
		return nil, status.Error(codes.Internal, ErrInternal.Error())
	}

	return &empty.Empty{}, nil
}

func (s serverApi) DeactivateClub(ctx context.Context, req *clubv1.DeactivateClubRequest) (*empty.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (s serverApi) UpdateClub(ctx context.Context, req *clubv1.UpdateClubRequest) (*clubv1.ClubObject, error) {
	err := validation.ValidateStruct(req,
		validation.Field(&req.UserId, validation.Required, validation.Min(1)),
		validation.Field(&req.ClubId, validation.Required, validation.Min(1)),
		validation.Field(&req.Name, validation.Length(domain.MinClubNameLen, domain.MaxClubNameLen)),
		validation.Field(&req.Description, validation.Length(domain.MinClubDescriptionLen, domain.MaxClubDescriptionLen)),
		validation.Field(&req.ClubType, validation.Length(domain.MinClubTypeLen, domain.MaxClubTypeLen)),
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	isAuthorized, err := s.permission.CanManageClub(ctx, req.GetClubId(), req.GetUserId())
	if err != nil {
		if errors.Is(err, domain.ErrUserNotClubMember) {
			return nil, status.Error(codes.PermissionDenied, ErrUserNotClubMember.Error())
		}
		return nil, status.Error(codes.Internal, ErrInternal.Error())
	}
	if !isAuthorized {
		return nil, status.Error(codes.PermissionDenied, ErrUserNonAuthorized.Error())
	}

	club, err := s.info.GetClub(ctx, req.GetClubId())
	if err != nil {
		if errors.Is(err, management.ErrClubNotExists) {
			return nil, status.Error(codes.NotFound, ErrClubNotFound.Error())
		}
		return nil, status.Error(codes.Internal, ErrInternal.Error())
	}

	paths := req.GetUpdateMask().GetPaths()
	for _, path := range paths {
		switch path {
		case "name":
			club.Name = req.GetName()
		case "description":
			club.Description = req.GetDescription()
		case "club_type":
			club.ClubType = req.GetClubType()
		}
	}

	err = s.management.UpdateClub(ctx, club)
	if err != nil {
		if errors.Is(err, management.ErrEditConflict) {
			return nil, status.Error(codes.Aborted, ErrEditConflict.Error())
		}
		return nil, status.Error(codes.Internal, ErrInternal.Error())
	}

	return club.ToClubObject(), nil

}

func (s serverApi) UpdateLogo(ctx context.Context, req *clubv1.UpdateLogoRequest) (*clubv1.UpdateLogoResponse, error) {
	err := validation.ValidateStruct(req,
		validation.Field(&req.ClubId, validation.Required, validation.Min(1)),
		validation.Field(&req.UserId, validation.Required, validation.Min(1)),
		validation.Field(&req.LogoUrl, validation.Required),
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	isAuthorized, err := s.permission.CanManageClub(ctx, req.GetClubId(), req.GetUserId())
	if err != nil {
		if errors.Is(err, domain.ErrUserNotClubMember) {
			return nil, status.Error(codes.PermissionDenied, ErrUserNotClubMember.Error())
		}
		return nil, status.Error(codes.Internal, ErrInternal.Error())
	}
	if !isAuthorized {
		return nil, status.Error(codes.PermissionDenied, ErrUserNonAuthorized.Error())
	}

	res, prevLogoUrl, err := s.management.UpdateLogo(ctx, req.GetClubId(), req.GetLogoUrl())
	if err != nil {
		switch {
		case errors.Is(err, management.ErrClubNotExists):
			return nil, status.Error(codes.NotFound, ErrInternal.Error())
		case errors.Is(err, management.ErrEditConflict):
			return nil, status.Error(codes.Aborted, ErrEditConflict.Error())
		default:
			return nil, status.Error(codes.Internal, ErrInternal.Error())
		}
	}

	return &clubv1.UpdateLogoResponse{
		Club:        res.ToClubObject(),
		PrevLogoUrl: prevLogoUrl,
	}, nil
}

func (s serverApi) UpdateBanner(ctx context.Context, req *clubv1.UpdateBannerRequest) (*clubv1.UpdateBannerResponse, error) {
	err := validation.ValidateStruct(req,
		validation.Field(&req.ClubId, validation.Required, validation.Min(1)),
		validation.Field(&req.UserId, validation.Required, validation.Min(1)),
		validation.Field(&req.BannerUrl, validation.Required),
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	isAuthorized, err := s.permission.CanManageClub(ctx, req.GetClubId(), req.GetUserId())
	if err != nil {
		if errors.Is(err, domain.ErrUserNotClubMember) {
			return nil, status.Error(codes.PermissionDenied, ErrUserNotClubMember.Error())
		}
		return nil, status.Error(codes.Internal, ErrInternal.Error())
	}
	if !isAuthorized {
		return nil, status.Error(codes.PermissionDenied, ErrUserNonAuthorized.Error())
	}

	res, prevBannerUrl, err := s.management.UpdateBanner(ctx, req.GetClubId(), req.GetBannerUrl())
	if err != nil {
		switch {
		case errors.Is(err, management.ErrClubNotExists):
			return nil, status.Error(codes.NotFound, ErrInternal.Error())
		case errors.Is(err, management.ErrEditConflict):
			return nil, status.Error(codes.Aborted, ErrInternal.Error())
		default:
			return nil, status.Error(codes.Internal, ErrInternal.Error())
		}
	}

	return &clubv1.UpdateBannerResponse{
		Club:          res.ToClubObject(),
		PrevBannerUrl: prevBannerUrl,
	}, nil
}

func (s serverApi) TransferOwnership(ctx context.Context, req *clubv1.TransferOwnershipRequest) (*empty.Empty, error) {
	err := validation.ValidateStruct(req,
		validation.Field(&req.ClubId, validation.Required, validation.Min(1)),
		validation.Field(&req.UserId, validation.Required, validation.Min(1), validation.NotIn(req.GetTargetId())),
		validation.Field(&req.TargetId, validation.Required, validation.Min(1)),
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = s.management.TransferOwnership(ctx, req.GetClubId(), req.GetUserId(), req.GetTargetId())
	if err != nil {
		switch {
		case errors.Is(err, management.ErrClubNotExists):
			return nil, status.Error(codes.NotFound, ErrClubNotFound.Error())
		case errors.Is(err, domain.ErrUserNotClubMember):
			return nil, status.Error(codes.NotFound, ErrUserNotClubMember.Error())
		case errors.Is(err, domain.ErrUserAlreadyClubOwner):
			return nil, status.Error(codes.AlreadyExists, domain.ErrUserAlreadyClubOwner.Error())
		case errors.Is(err, domain.ErrUserNotClubOwner):
			return nil, status.Error(codes.PermissionDenied, domain.ErrUserNotClubOwner.Error())
		default:
			return nil, status.Error(codes.Internal, ErrInternal.Error())
		}
	}

	return &empty.Empty{}, nil
}
