package dtos

type UpdateNodeDTO struct {
	Id              string          `json:"id"`
	Name            string          `json:"name" binding:"required"`
	OperatingSystem OperatingSystem `json:"operatingSystem" binding:"required,operatingsystem"`
}
