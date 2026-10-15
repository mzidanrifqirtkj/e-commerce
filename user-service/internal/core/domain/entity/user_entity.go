package entity

type UserEntity struct {
	ID         int64
	Name       string
	Email      string
	Password   string
	Address    string
	Phone      string
	Photo      string
	Lat        string
	Lng        string
	IsVerified bool
	RoleName   string
	Token      string
}

type QueryStringCustomer struct {
	Search    string
	Page      int64
	Limit     int64
	OrderBy   string
	OrderType string
}
