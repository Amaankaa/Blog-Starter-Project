package usecases

import (
	"github.com/Amaankaa/Blog-Starter-Project/Domain/user"
)

type UserUsecase struct {
	userRepo userpkg.UserRepository
}

func NewUserUsecase(userRepo userpkg.UserRepository) *UserUsecase {
	return &UserUsecase{
		userRepo: userRepo,
	}
}

func (uu *UserUsecase) IRegisterUser(user userpkg.User) (userpkg.User, error) {
	return uu.userRepo.RegisterUser(user)
}