package dtos

type ChangeEmail struct {
	Email *string `json:"email" binding:"required,email"`
}

type UserDto struct{}
