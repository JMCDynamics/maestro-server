package dtos

type CreateUserDTO struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
