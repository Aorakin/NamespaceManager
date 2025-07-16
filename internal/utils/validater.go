package utils

import "github.com/go-playground/validator/v10"

func CheckValidater(i interface{}) error {
	validate := validator.New()
	if err := validate.Struct(i); err != nil {
		return err
	}
	return nil
}
