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
	Club Club
	User User
}

func MapClubUserArrToClubList(cu []*ClubUser) []*clubv1.NotActivatedClubsList {
	clubUserObjects := make([]*clubv1.NotActivatedClubsList, len(cu))
	for i, clubUser := range cu {
		clubUserObjects[i] = &clubv1.NotActivatedClubsList{
			Clubs: ClubToClubObject(&clubUser.Club),
			Owner: UserToUserObject(&clubUser.User),
		}
	}
	return clubUserObjects
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
