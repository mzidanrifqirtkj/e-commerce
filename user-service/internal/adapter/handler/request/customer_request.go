package request

type CustomerRequest struct {
	Name                 string  `json:"name" validate:"required"`
	Email                string  `json:"email" validate:"email,required"`
	Password             string  `json:"password" validate:"min=8,required"`
	PasswordConfirmation string  `json:"password_confirmation" validate:"min=8,required"`
	Phone                int64   `json:"phone" validate:"required,number"`
	Address              string  `json:"address"`
	Photo                string  `json:"photo"`
	Lat                  float64 `json:"lat"`
	Lng                  float64 `json:"lng"`
	RoleID               int64   `json:"role_id" validate:"required"`
}

// UpdateCustomerRequest: password is optional, empty means keep the current one.
type UpdateCustomerRequest struct {
	Name                 string  `json:"name" validate:"required"`
	Email                string  `json:"email" validate:"email,required"`
	Password             string  `json:"password" validate:"omitempty,min=8"`
	PasswordConfirmation string  `json:"password_confirmation" validate:"omitempty,min=8"`
	Phone                int64   `json:"phone" validate:"required,number"`
	Address              string  `json:"address"`
	Photo                string  `json:"photo"`
	Lat                  float64 `json:"lat"`
	Lng                  float64 `json:"lng"`
	RoleID               int64   `json:"role_id" validate:"required"`
}
