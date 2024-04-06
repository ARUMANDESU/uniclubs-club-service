package domain

import (
	"fmt"
	clubv1 "github.com/ARUMANDESU/uniclubs-protos/gen/go/club"
	"time"
)

type Role struct {
	ID          int64
	ClubID      int64
	Name        string
	Permissions Permissions
	Position    int32
	Color       int32
	UpdatedAt   time.Time
}

func (r *Role) ToRoleProto() *clubv1.Role {
	_ = r.Permissions.HexToStringArr()

	return &clubv1.Role{
		Id:          r.ID,
		Name:        r.Name,
		Permissions: r.Permissions.PermissionsArr,
		Position:    r.Position,
		Color:       r.Color,
	}
}

func MapToRoleObjectArr(r []Role) []*clubv1.Role {
	roles := make([]*clubv1.Role, len(r))
	for i, role := range r {
		roles[i] = role.ToRoleProto()
	}

	return roles
}

func GetHighestRolePosition(roles []Role) (int32, error) {
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
