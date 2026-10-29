package response

// CategoryResponse is a row of the admin category list: it carries the product
// count but not the description.
type CategoryResponse struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Icon         string `json:"icon"`
	Slug         string `json:"slug"`
	Status       bool   `json:"status"`
	TotalProduct int64  `json:"total_product"`
}

// CategoryDetailResponse is a single category: it carries the description but
// not the product count.
type CategoryDetailResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Icon        string `json:"icon"`
	Slug        string `json:"slug"`
	Status      bool   `json:"status"`
	Description string `json:"description"`
}
