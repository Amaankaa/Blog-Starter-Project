package usecases

import (
	"github.com/Amaankaa/Blog-Starter-Project/Domain"
)

type UserUsecase struct {
	userRepo domain.UserRepository
}

func NewUserUsecase(userRepo domain.UserRepository) *UserUsecase {
	return &UserUsecase{
		userRepo: userRepo,
	}
}

func (uu *UserUsecase) RegisterUser(user domain.User) (domain.User, error) {
	return uu.userRepo.RegisterUser(user)
}