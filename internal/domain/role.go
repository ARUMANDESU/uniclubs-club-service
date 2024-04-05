package domain

import (
	"fmt"
	clubv1 "github.com/ARUMANDESU/uniclubs-protos/gen/go/club"
)

type Role struct {
	ID          int
	Name        string
	Permissions Permissions
	Position    int
	Color       int
}

func MapToRoleObjectArr(r []Role) []*clubv1.Role {
	roles := make([]*clubv1.Role, len(r))
	for i, role := range r {
		role.Permissions.HexToStringArr()
		roles[i] = &clubv1.Role{
			Name:        role.Name,
			Permissions: role.Permissions.PermissionsArr,
		}
	}

	return roles
}

func GetHighestRolePosition(roles []Role) (int, error) {
	const op = "domain.role.GetHighestPositionRole"
	if len(roles) == 0 {
		return 0, fmt.Errorf("%s: no roles provided", op)
	}

	highestRole := roles[0]
	for _, role := range roles {
		if role.Position > highestRole.Position {
			highestRole = role
		}
	}
	return highestRole.Position, nil
}
