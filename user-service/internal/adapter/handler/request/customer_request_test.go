package request

import (
	"testing"

	"github.com/go-playground/validator/v10"
)

// UpdateCustomer must accept a payload without a password (keep the current
// hash) and still reject one that is present but too short.
func TestUpdateCustomerRequestPasswordOptional(t *testing.T) {
	v := validator.New()

	base := UpdateCustomerRequest{
		Name:   "Budi",
		Email:  "budi@example.com",
		Phone:  628123456789,
		RoleID: 2,
	}

	if err := v.Struct(base); err != nil {
		t.Fatalf("empty password rejected: %v", err)
	}

	short := base
	short.Password = "123"
	if err := v.Struct(short); err == nil {
		t.Fatal("3-character password accepted")
	}

	ok := base
	ok.Password = "password123"
	ok.PasswordConfirmation = "password123"
	if err := v.Struct(ok); err != nil {
		t.Fatalf("valid password rejected: %v", err)
	}
}
