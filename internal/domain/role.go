package domain

import "fmt"

type Role struct {
	ID          int
	Name        string
	Permissions Permissions
	Position    int
	Color       int
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
