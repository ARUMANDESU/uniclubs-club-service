package domain

import clubv1 "github.com/ARUMANDESU/uniclubs-protos/gen/go/club"

type User struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Barcode   string `json:"barcode"`
	AvatarURL string `json:"avatar_url"`
	Role      string `json:"role"`
}

func UserToUserObject(user *User) *clubv1.UserObject {
	return &clubv1.UserObject{
		UserId:    user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Barcode:   user.Barcode,
		AvatarUrl: user.AvatarURL,
		Role:      user.Role,
	}
}
