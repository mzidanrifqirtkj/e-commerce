package response

type SignInResponse struct {
	AccessToken string `json:"access_token"`
	Role        string `json:"role"`
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Lat         string `json:"lat"`
	Lng         string `json:"lng"`
	Email       string `json:"email"` // Added missing field
	Phone       string `json:"phone"` // Added missing field
}

type ProfileResponse struct {
	RoleName string `json:"role"`
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Lat      string `json:"lat"`
	Lng      string `json:"lng"`
	Email    string `json:"email"` // Added missing field
	Phone    string `json:"phone"` // Added missing field
	Address  string `json:"address"`
	Photo    string `json:"photo"`
}

type CustomerResponse struct {
	RoleName string `json:"role,omitempty"`
	RoleID   int64  `json:role_id`
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Lat      string `json:"lat"`
	Lng      string `json:"lng"`
	Email    string `json:"email"` // Added missing field
	Phone    string `json:"phone"` // Added missing field
	Address  string `json:"address"`
	Photo    string `json:"photo"`
}
