package model

import "gopkg.in/go-playground/validator.v9"

var validate *validator.Validate

func GetValidator() *validator.Validate {
	if validate == nil {
		validate = validator.New()
	}
	return validate
}
