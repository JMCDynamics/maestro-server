package dtos

import "github.com/go-playground/validator/v10"

type CreateNodeDTO struct {
	Name string `json:"name" binding:"required"`
	OperatingSystem OperatingSystem `json:"operatingSystem" binding:"required,operatingsystem"` 
}

func ValidateOperatingSystem(fl validator.FieldLevel) bool {
	allowedSystems := []OperatingSystem{WINDOWS, LINUX}
	operatingSystem := fl.Field().String()

	for _, os := range allowedSystems {
		if operatingSystem == string(os) {
			return true
		}
	}

	return false
}