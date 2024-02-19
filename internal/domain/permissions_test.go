package domain

import (
	"testing"
)

func TestPermissions_HexToStringArr(t *testing.T) {
	tests := []struct {
		name     string
		hex      uint64
		expected []string
	}{
		{"No permissions", 0x0, []string{}},
		{"Administrator", 0x1, []string{"Administrator"}},
		{"2 permissions: ManageClub, ManageMembership ", 0x6, []string{"ManageClub", "ManageMembership"}},
		{"KickMember", 0x8, []string{"KickMember"}}, // Assuming 0x8 is the bit for AddMembers
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Permissions{PermissionsHex: tt.hex}
			err := p.HexToStringArr()
			if err != nil {
				t.Fatalf("HexToStringArr() error = %v", err)
			}
			if len(p.PermissionsArr) != len(tt.expected) {
				t.Fatalf("Expected %v, got %v", tt.expected, p.PermissionsArr)

			}
			slicesEqual(t, tt.expected, p.PermissionsArr)

		})
	}
}

func TestStringArrToHex(t *testing.T) {
	tests := []struct {
		name        string
		permissions []string
		expected    uint64
	}{
		{"No permissions", []string{}, 0x0},
		{"Administrator", []string{"Administrator"}, 0x1},
		{"2 permissions: ManageClub, ManageMembership", []string{"ManageClub", "ManageMembership"}, 0x6},
		{"KickMember", []string{"KickMember"}, 0x8}, // Assuming 0x8 is the bit for AddMembers
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Permissions{PermissionsArr: tt.permissions}
			err := p.StringArrToHex()
			if err != nil {
				t.Fatalf("StringArrToHex() error = %v", err)
			}
			if p.PermissionsHex != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, p.PermissionsHex)
			}
		})

	}
}

func TestAccumulatePermissions(t *testing.T) {
	tests := []struct {
		name  string
		roles []Role
		want  uint64
	}{
		{"NoPermissions", []Role{}, 0},
		{"SinglePermission", []Role{{Permissions: Permissions{PermissionsHex: ManageClub}}}, ManageClub},
		{"MultiplePermissions", []Role{
			{Permissions: Permissions{PermissionsHex: ManageClub}},
			{Permissions: Permissions{PermissionsHex: ManageMembership}},
			{Permissions: Permissions{PermissionsHex: KickMember}},
		},
			ManageClub | ManageMembership | KickMember},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AccumulatePermissions(tt.roles); got != tt.want {
				t.Errorf("AccumulatePermissions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasPermission(t *testing.T) {
	tests := []struct {
		name            string
		userPermissions uint64
		permission      uint64
		want            bool
	}{
		{"HasPermission", ManageRoles | BanMember, BanMember, true},
		{"DoesNotHavePermission", ManageRoles, ManageMembership, false},
		{"MultiplePermissionsTrue", ManageClub | BanMember, ManageClub, true},
		{"MultiplePermissionsFalse", ManageRoles | KickMember, BanMember, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasPermission(tt.userPermissions, tt.permission); got != tt.want {
				t.Errorf("HasPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}

func slicesEqual(t *testing.T, expected, got []string) {
	t.Helper()

	if len(expected) != len(got) {
		t.Errorf("Expected %v, got %v", expected, got)
	}
	m := make(map[string]bool, len(expected))
	for _, s := range expected {
		m[s] = true
	}

	for _, s := range got {
		if !m[s] {
			t.Errorf("Expected %v, got %v", expected, got)
		}
	}

}
