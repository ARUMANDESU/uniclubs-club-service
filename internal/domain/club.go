package domain

import "time"

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
