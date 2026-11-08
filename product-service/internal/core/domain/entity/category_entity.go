package entity

type CategoryEntity struct {
	ID          int64
	ParentID    *int64
	Name        string
	Icon        string
	Status      bool
	Slug        string
	Description string
	Products    []ProductEntity
}

type QueryStringEntity struct {
	Search    string
	Page      int
	Limit     int
	OrderBy   string
	OrderType string
	// Status filters by the published flag. nil = no filter (all rows);
	// &true = published only, &false = unpublished only.
	Status *bool
}
