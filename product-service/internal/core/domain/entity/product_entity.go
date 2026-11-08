package entity

type ProductEntity struct {
	ID           int64
	CategorySlug string
	ParentID     *int64
	Name         string
	Image        string
	Description  string
	RegularPrice float64
	SalePrice    float64
	Unit         string
	Weight       int
	Stock        int
	Variant      int
	Status       string
	CategoryName string
}

type QueryStringProduct struct {
	Search       string
	Page         int
	Limit        int
	OrderBy      string
	OrderType    string
	CategorySlug string
	StartPrice   int64
	EndPrice     int64
}
