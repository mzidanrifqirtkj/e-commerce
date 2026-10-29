package request

// CategoryRequest is the payload for creating or updating a category.
//
// Status is a pointer so "not sent" stays distinguishable from "explicitly
// false": Create defaults it to true, matching the DB column default.
type CategoryRequest struct {
	ParentID    *int64 `json:"parent_id"`
	Name        string `json:"name" validate:"required"`
	Icon        string `json:"icon" validate:"required"`
	Slug        string `json:"slug"`
	Status      *bool  `json:"status"`
	Description string `json:"description"`
}
