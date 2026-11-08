package request

// CategoryRequest is the payload for creating or updating a category.
//
// Status is a pointer so "not sent" stays distinguishable from "explicitly
// false": Create defaults it to published, matching the DB column default.
// The wire value is the word "published" or "unpublished", not a boolean.
type CategoryRequest struct {
	ParentID    *int64  `json:"parent_id"`
	Name        string  `json:"name" validate:"required"`
	Icon        string  `json:"icon" validate:"required"`
	Slug        string  `json:"slug"`
	Status      *string `json:"status" validate:"omitempty,oneof=published unpublished"`
	Description string  `json:"description"`
}
