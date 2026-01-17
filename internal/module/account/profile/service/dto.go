// internal/module/account/profile/dto.go
package profile_service

type UpdateProfileInput struct {
	Name string `validate:"required"`
	// 1. json tag "imageUrl" (agar mapping sesuai)
	// 2. validate "omitempty,url"
	//    Artinya: Jika kosong/null -> Skip validasi. Jika ada isi -> Cek format URL.
	ImageUrl    *string `json:"imageUrl" validate:"omitempty,url"`
	CountryCode string  `validate:"required,max=5"`
	Phone       *string `validate:"required,max=20"`
	Description *string `validate:"max=255"`
}

type UpdatePasswordInput struct {
	OldPassword string `validate:"required"`
	NewPassword string `validate:"required,min=8,max=20"`
}

type SetupPasswordInput struct {
	Password string `validate:"required,min=8,max=20"`
	From     string `validate:"required,url"`
}
