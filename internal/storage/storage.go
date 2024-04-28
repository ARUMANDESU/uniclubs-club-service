package storage

import "errors"

var (
	ErrUserExists          = errors.New("user already exists")
	ErrUserNotExists       = errors.New("user does not exists")
	ErrClubNotExists       = errors.New("club does not exists")
	ErrClubOrRoleNotExists = errors.New("club or role does not exists")
	ErrUserNotClubMember   = errors.New("user is not club member")
	ErrEditConflict        = errors.New("edit conflict")
	ErrBanRecordNotExists  = errors.New("ban record does not exists")
)
