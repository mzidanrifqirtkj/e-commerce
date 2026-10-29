// Package validator adapts go-playground/validator to echo's Validator
// interface, returning the first translated message as a plain error.
package validator

import (
	"errors"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

type Validator struct {
	Validator  *validator.Validate
	Translator ut.Translator
}

func NewValidator() *Validator {
	enLocale := en.New()
	uni := ut.New(enLocale, enLocale)
	trans, found := uni.GetTranslator("en")
	if !found {
		panic("[NewValidator-1] translator not found")
	}

	return &Validator{
		Validator:  validator.New(),
		Translator: trans,
	}
}

func (v *Validator) Validate(i interface{}) error {
	err := v.Validator.Struct(i)
	if err == nil {
		return nil
	}

	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) || len(validationErrors) == 0 {
		return err
	}

	first := validationErrors[0]
	return errors.New(first.Field() + ": " + first.Translate(v.Translator))
}
