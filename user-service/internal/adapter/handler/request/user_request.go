package request

type SigInRequest struct {
	Email    string `json:"email" validate:"email, required"`
	Password string `json:"password" validate:"min=8, required"`
}
