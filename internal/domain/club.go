package domain

import (
	clubv1 "github.com/ARUMANDESU/uniclubs-protos/gen/go/club"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type Club struct {
	ID           int64
	Name         string
	Description  string
	ClubType     string
	LogoURL      string
	BannerURL    string
	NumOFMembers int
	CreatedAt    time.Time
}

type ClubUser struct {
	Club
	User
}

func ClubToClubObject(club *Club) *clubv1.ClubObject {
	return &clubv1.ClubObject{
		ClubId:          club.ID,
		Name:            club.Name,
		Description:     club.Description,
		ClubType:        club.ClubType,
		LogoUrl:         club.LogoURL,
		BannerUrl:       club.BannerURL,
		NumberOfMembers: int64(club.NumOFMembers),
		CreatedAt:       timestamppb.New(club.CreatedAt),
	}
}

func MapClubArrToClubObjectArr(clubs []*Club) []*clubv1.ClubObject {
	clubObjects := make([]*clubv1.ClubObject, len(clubs))
	for i, user := range clubs {
		clubObjects[i] = ClubToClubObject(user)
	}
	return clubObjects
}
