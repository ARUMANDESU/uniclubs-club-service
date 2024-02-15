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
	NumOFMembers int64
	CreatedAt    time.Time
	Roles        []Role
}

type Role struct {
	ID          int
	Name        string
	Permissions []string
}

func (c Club) ToClubObject() *clubv1.ClubObject {
	roles := make([]*clubv1.Role, len(c.Roles))
	for i, role := range c.Roles {
		roles[i] = &clubv1.Role{
			Name:        role.Name,
			Permissions: role.Permissions,
		}
	}

	return &clubv1.ClubObject{
		ClubId:          c.ID,
		Name:            c.Name,
		Description:     c.Description,
		ClubType:        c.ClubType,
		LogoUrl:         c.LogoURL,
		BannerUrl:       c.BannerURL,
		NumberOfMembers: c.NumOFMembers,
		CreatedAt:       timestamppb.New(c.CreatedAt),
		Roles:           roles,
	}
}

type ClubUser struct {
	Club Club
	User User
}

func MapClubUserArrToClubList(cu []*ClubUser) []*clubv1.NotActivatedClubsList {
	clubUserObjects := make([]*clubv1.NotActivatedClubsList, len(cu))
	for i, clubUser := range cu {
		clubUserObjects[i] = &clubv1.NotActivatedClubsList{
			Clubs: clubUser.Club.ToClubObject(),
			Owner: clubUser.User.ToUserObject(),
		}
	}
	return clubUserObjects
}

func MapClubArrToClubObjectArr(clubs []*Club) []*clubv1.ClubObject {
	clubObjects := make([]*clubv1.ClubObject, len(clubs))
	for i, user := range clubs {
		clubObjects[i] = user.ToClubObject()
	}
	return clubObjects
}
