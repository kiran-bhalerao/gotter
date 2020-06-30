package utils

import "github.com/asaskevich/govalidator"

func Validator(inputs interface{}) error {
	_, err := govalidator.ValidateStruct(inputs)
	return err
}
