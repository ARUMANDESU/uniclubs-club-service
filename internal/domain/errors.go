package domain

import "errors"

var (
	/*	ErrInternal              = errors.New("internal error")
		ErrUserExists            = errors.New("user already exists")
		ErrUserNotExists         = errors.New("user does not exists")
		ErrClubNotExists         = errors.New("club does not exists")
		ErrClubOrRoleNotExists   = errors.New("club or role may not exists")
		ErrUserNotClubMember     = errors.New("user is not club member")
		ErrEditConflict          = errors.New("unable to update the record due to an edit conflict, please try again")
		ErrUserNonAuthorized     = errors.New("user does not have permission")*/
	ErrUserAlreadyRoleMember    = errors.New("user is already a role member")
	ErrMemberNotFound           = errors.New("member not found")
	ErrMemborNotHavePermissions = errors.New("member does not have permissions")
)
