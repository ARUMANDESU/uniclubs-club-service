package domain

import (
	clubv1 "github.com/ARUMANDESU/uniclubs-protos/gen/go/club"
	"time"
)

type BanRecord struct {
	ID       int64
	ClubID   int64
	User     User
	Admin    User
	Reason   string
	BannedAt time.Time
}

func BanRecordToBanRecordObject(banRecord *BanRecord) *clubv1.BanRecord {
	return &clubv1.BanRecord{
		Id:       banRecord.ID,
		ClubId:   banRecord.ClubID,
		User:     banRecord.User.ToUserObject(),
		Admin:    banRecord.Admin.ToUserObject(),
		Reason:   banRecord.Reason,
		BannedAt: banRecord.BannedAt.String(),
	}
}

func BanRecordsToBanRecordObjects(banRecords []*BanRecord) []*clubv1.BanRecord {
	var banRecordObjects []*clubv1.BanRecord
	for _, banRecord := range banRecords {
		banRecordObjects = append(banRecordObjects, BanRecordToBanRecordObject(banRecord))
	}
	return banRecordObjects
}
