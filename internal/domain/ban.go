package domain

import "time"

type BanRecord struct {
	ID       int64
	ClubID   int64
	UserID   int64
	AdminID  int64
	Reason   string
	BannedAt time.Time
}
