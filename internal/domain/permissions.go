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

func (p *Permissions) StringArrToHex() error {
	const op = "domain.permission.StringArrToHex"

	var bitValue uint64 = 0

	for _, perm := range p.PermissionsArr {
		if val, ok := Values[perm]; ok {
			bitValue |= val
		} else {
			return fmt.Errorf("%s: invalid permission name: %s", op, perm)
		}
	}

	p.PermissionsHex = bitValue
	return nil
}

func AccumulatePermissions(roles []Role) (accumulatedPermissions uint64) {
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
