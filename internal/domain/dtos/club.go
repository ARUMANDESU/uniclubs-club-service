package dtos

import clubv1 "github.com/ARUMANDESU/uniclubs-protos/gen/go/club"

type CreateClubDTO struct {
	Name        string
	Description string
	ClubType    string
	OwnerID     int64
}

type CreateRoleDTO struct {
	ClubID int64
	UserID int64
	Name   string
	Color  int32
}

type ChangeRolesPositionDTO struct {
	RoleID   int64
	Position int32
}

type BanMemberDTO struct {
	ClubID  int64
	UserID  int64
	AdminID int64
	Reason  string
}

type UnbanUserDTO struct {
	ClubID  int64
	UserID  int64
	AdminID int64
}

func MapToChangeRolesPositionDTOArr(roles []*clubv1.ChangeRolesPositionItems) []*ChangeRolesPositionDTO {
	dto := make([]*ChangeRolesPositionDTO, len(roles))
	for i, role := range roles {
		dto[i] = &ChangeRolesPositionDTO{
			RoleID:   role.GetId(),
			Position: role.GetPosition(),
		}
	}

	return dto
}

func CreateClubRequestToDTO(req *clubv1.CreateClubRequest) CreateClubDTO {
	return CreateClubDTO{
		Name:        req.Name,
		Description: req.Description,
		ClubType:    req.ClubType,
		OwnerID:     req.OwnerId,
	}
}

func BanMemberRequestToDTO(req *clubv1.BanMemberFromClubRequest) BanMemberDTO {
	return BanMemberDTO{
		ClubID:  req.GetClubId(),
		UserID:  req.GetTargetId(),
		AdminID: req.GetUserId(),
		Reason:  req.GetReason(),
	}
}

func UnbanUserRequestToDTO(req *clubv1.UnbanUserFromClubRequest) UnbanUserDTO {
	return UnbanUserDTO{
		ClubID:  req.GetClubId(),
		UserID:  req.GetTargetId(),
		AdminID: req.GetUserId(),
	}
}
