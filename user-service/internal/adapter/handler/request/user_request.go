package request

type SigInRequest struct {
	Email    string `json:"email" validate:"email,required"`
	Password string `json:"password" validate:"min=8,required"`
}

type SignUpRequest struct {
	Name                 string `json:"name" validate:"required"`
	Email                string `json:"email" validate:"email,required"`
	Password             string `json:"password" validate:"min=8,required"`
	PasswordConfirmation string `json:"password_confirmation" validate:"min=8,required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"email,required"`
}

type UpdatePasswordRequest struct {
	NewPassword     string `json:"password_new" validate:"required"`
	ConfirmPassword string `json:"password_confirmation" validate:"required"`
}

type UpdateDataUserRequest struct {
	Name    string `json:"name" validate:"required"`
	Email   string `json:"email" validate:"email,required"`
	Phone   int64  `json:"phone" validate:"required,number"`
	Address string `json:"address" validate:"required"`
	Photo   string `json:"photo" validate:"required"`
	Lat     string `json:"lat" validate:"required"`
	Lng     string `json:"lng" validate:"required"`
}
