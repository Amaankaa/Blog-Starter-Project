package usecases

import (
	"context"
	"errors"

	userpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/user"
	utils "github.com/Amaankaa/Blog-Starter-Project/Domain/utils"
	"github.com/Amaankaa/Blog-Starter-Project/Domain/services"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserUsecase struct {
	userRepo       userpkg.IUserRepository
	passwordSvc    userpkg.IPasswordService
	emailVerifier  services.IEmailVerifier
}

func NewUserUsecase(
	userRepo userpkg.IUserRepository,
	passwordSvc userpkg.IPasswordService,
	emailVerifier services.IEmailVerifier,
) *UserUsecase {
	return &UserUsecase{
		userRepo:      userRepo,
		passwordSvc:   passwordSvc,
		emailVerifier: emailVerifier,
	}
}

func (uu *UserUsecase) RegisterUser(ctx context.Context, user userpkg.User) (userpkg.User, error) {
	// Basic field validation
	if user.Username == "" || user.Email == "" || user.Password == "" || user.Fullname == "" {
		return userpkg.User{}, errors.New("all fields are required")
	}

	// Email format
	if !utils.IsValidEmail(user.Email) {
		return userpkg.User{}, errors.New("invalid email format")
	}

	// Real email check
	isReal, err := uu.emailVerifier.IsRealEmail(user.Email)
	if err != nil {
		return userpkg.User{}, errors.New("failed to verify email: " + err.Error())
	}
	if !isReal {
		return userpkg.User{}, errors.New("email is unreachable")
	}

	// Strong password check
	if !utils.IsStrongPassword(user.Password) {
		return userpkg.User{}, errors.New("password must be at least 8 chars, with upper, lower, number, and special char")
	}

	// Username and email uniqueness
	exists, _ := uu.userRepo.ExistsByUsername(ctx, user.Username)
	if exists {
		return userpkg.User{}, errors.New("username already taken")
	}
	exists, _ = uu.userRepo.ExistsByEmail(ctx, user.Email)
	if exists {
		return userpkg.User{}, errors.New("email already taken")
	}

	// Assign role
	count, err := uu.userRepo.CountUsers(ctx)
	if err != nil {
		return userpkg.User{}, err
	}
	if count == 0 {
		user.Role = "admin"
	} else {
		user.Role = "user"
	}

	// Password hashing
	hashed, err := uu.passwordSvc.HashPassword(user.Password)
	if err != nil {
		return userpkg.User{}, err
	}
	user.Password = hashed
	user.ID = primitive.NewObjectID()
	user.IsVerified = false

	_, err = uu.userRepo.CreateUser(ctx, user)
	if err != nil {
		return userpkg.User{}, err
	}

	user.Password = "" // scrub before return
	return user, nil
}