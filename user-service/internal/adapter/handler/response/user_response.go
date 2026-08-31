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
