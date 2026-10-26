package entity

type CategoryEntity struct {
	ID          int64
	ParentID    *int64
	Name        string
	Icon        string
	Status      string
	Slug        string
	Description string
	Products    []ProductEntity
}
