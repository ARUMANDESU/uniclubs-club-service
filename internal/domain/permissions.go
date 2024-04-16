package domain

import "fmt"

const (
	Administrator uint64 = 1 << iota
	ManageClub
	ManageMembership
	KickMember
	BanMember
	ManageRoles
	ALL = Administrator | ManageRoles | ManageMembership | KickMember | BanMember | ManageClub
)

var Names = map[uint64]string{
	Administrator:    "Administrator",
	ManageClub:       "ManageClub",
	ManageMembership: "ManageMembership",
	KickMember:       "KickMember",
	BanMember:        "BanMember",
	ManageRoles:      "ManageRoles",
}

var Values = map[string]uint64{
	"Administrator":    Administrator,
	"ManageClub":       ManageClub,
	"ManageMembership": ManageMembership,
	"KickMember":       KickMember,
	"BanMember":        BanMember,
	"ManageRoles":      ManageRoles,
}

var PermissionList = []string{
	"Administrator",
	"ManageClub",
	"ManageMembership",
	"KickMember",
	"BanMember",
	"ManageRoles",
}

type Permissions struct {
	PermissionsHex uint64
	PermissionsArr []string
}

func (p *Permissions) HexToStringArr() error {
	var permissions []string
	// Iterate over all possible permissions
	for bit, name := range Names {
		// Check if the permission bit is set
		if p.PermissionsHex&bit != 0 {
			permissions = append(permissions, name)
		}
	}

	p.PermissionsArr = permissions
	return nil
}

func PermissionsHexToStringArr(p uint64) []string {
	var permissions []string
	// Iterate over all possible permissions
	for bit, name := range Names {
		// Check if the permission bit is set
		if p&bit != 0 {
			permissions = append(permissions, name)
		}
	}

	return permissions
}

func StringArrToHex(p []string) (uint64, error) {
	const op = "domain.permission.StringArrToHex"

	var bitValue uint64 = 0

	for _, perm := range p {
		if val, ok := Values[perm]; ok {
			bitValue |= val
		} else {
			return 0, fmt.Errorf("%s: invalid permission name: %s", op, perm)
		}
	}

	return bitValue, nil
}

func AccumulatePermissions(roles []*Role) (accumulatedPermissions uint64) {
	for _, role := range roles {
		if role.Permissions.PermissionsHex&Administrator == Administrator {
			return ALL
		}
		accumulatedPermissions |= role.Permissions.PermissionsHex
	}
	return accumulatedPermissions
}

func HasPermission(userPermissions uint64, permission uint64) bool {
	return userPermissions&permission != 0
}

func UserHasPermissions(userPermissions uint64, permissions uint64) (bool, error) {
	res := userPermissions&permissions == permissions
	if !res {
		doNotHavePermsArr := PermissionsHexToStringArr(userPermissions ^ permissions)
		return false, fmt.Errorf("%v: %w", doNotHavePermsArr, ErrMemborNotHavePermissions)
	}
	return res, nil
}
