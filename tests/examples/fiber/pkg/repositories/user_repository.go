package repositories

import "fiber/pkg/dtos"

type UserRepository struct{}

func (r *UserRepository) GetById(userId int) (*dtos.UserDto, error) {
	panic("not implemented")
}
