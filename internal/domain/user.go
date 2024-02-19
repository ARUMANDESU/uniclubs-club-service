package domain

import clubv1 "github.com/ARUMANDESU/uniclubs-protos/gen/go/club"

type User struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Barcode   string `json:"barcode"`
	AvatarURL string `json:"avatar_url"`
	Role      Role
}

func (u User) ToUserObject() *clubv1.UserObject {
	u.Role.Permissions.HexToStringArr()
	return &clubv1.UserObject{
		UserId:    u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Barcode:   u.Barcode,
		AvatarUrl: u.AvatarURL,
		Role: &clubv1.Role{
			Name:        u.Role.Name,
			Permissions: u.Role.Permissions.PermissionsArr,
		},
	}
}

func MapUserArrToUserObjectArr(users []*User) []*clubv1.UserObject {
	userObjects := make([]*clubv1.UserObject, len(users))
	for i, user := range users {
		userObjects[i] = user.ToUserObject()
	}
	return userObjects
}
