package usecases

import (
	"github.com/Amaankaa/Blog-Starter-Project/Domain/user"
)

type UserUsecase struct {
	userRepo userpkg.IUserRepository
}

func NewUserUsecase(userRepo userpkg.IUserRepository) *UserUsecase {
	return &UserUsecase{
		userRepo: userRepo,
	}
}

func (uu *UserUsecase) RegisterUser(user userpkg.User) (userpkg.User, error) {
	return uu.userRepo.RegisterUser(user)
}