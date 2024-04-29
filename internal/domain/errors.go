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
	ErrUserAlreadyRoleMember                  = errors.New("user is already a role member")
	ErrUserAlreadyClubMember                  = errors.New("user is already a club member")
	ErrUserAlreadySentJoinRequest             = errors.New("user already sent join request to the club")
	ErrMemberNotFound                         = errors.New("member not found")
	ErrMemberNotHavePermissions               = errors.New("member does not have these permissions")
	ErrMemberNotHavePermissionsToEditRole     = errors.New("member does not have permissions to change role by id")
	ErrClubOrRoleNotExists                    = errors.New("club or role does not exists")
	ErrInsufficientRolePosition               = errors.New("user's highest role position is less than target's")
	ErrUserNotClubMember                      = errors.New("user is not club member")
	ErrTargetNotClubMember                    = errors.New("target user is not club member")
	ErrMemberNotHaveAccessToRemovePermissions = errors.New("member does not have permissions to remove role permissions that member does not have")
	ErrCannotEditRoleMember                   = errors.New("no one can edit role 'member'")
	ErrOwnerCannotLeaveClub                   = errors.New("owner cannot leave the club, first pass on the ownership to other member")
	ErrOwnerCannotBeKickedOut                 = errors.New("owner cannot be kicked out from the club")
	ErrOwnerCannotBeBanned                    = errors.New("owner cannot be banned from the club")
	ErrUserAlreadyClubOwner                   = errors.New("user is already club owner")
	ErrUserNotClubOwner                       = errors.New("user is not club owner")
	ErrUserAlreadyBanned                      = errors.New("user is already banned")
	ErrUserNotBanned                          = errors.New("user is not banned")
	ErrUserBanned                             = errors.New("user is banned")
)
