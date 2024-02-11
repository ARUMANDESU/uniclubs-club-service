package dtos

import clubv1 "github.com/ARUMANDESU/uniclubs-protos/gen/go/club"

type CreateClubDTO struct {
	Name        string
	Description string
	ClubType    string
	OwnerID     int64
}

func CreateClubRequestToDTO(req *clubv1.CreateClubRequest) CreateClubDTO {
	return CreateClubDTO{
		Name:        req.Name,
		Description: req.Description,
		ClubType:    req.ClubType,
		OwnerID:     req.OwnerId,
	}
}
