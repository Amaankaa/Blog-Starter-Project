// Filename: usecase_test.go
package usecases_test

import (
	"context"
	"testing"

	userpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/user"
	"github.com/Amaankaa/Blog-Starter-Project/Usecases"
	"github.com/Amaankaa/Blog-Starter-Project/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserUsecaseTestSuite struct {
	suite.Suite
	ctx              context.Context
	mockUserRepo     *mocks.IUserRepository
	mockPasswordSvc  *mocks.IPasswordService
	mockTokenRepo    *mocks.ITokenRepository
	mockJWTService   *mocks.IJWTService
	mockEmailVerifier *mocks.IEmailVerifier
	mockEmailSender  *mocks.IEmailSender
	mockResetRepo    *mocks.IPasswordResetRepository
	usecase          *usecases.UserUsecase
}

func TestUserUsecaseTestSuite(t *testing.T) {
	suite.Run(t, new(UserUsecaseTestSuite))
}

func (s *UserUsecaseTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.mockUserRepo = new(mocks.IUserRepository)
	s.mockPasswordSvc = new(mocks.IPasswordService)
	s.mockTokenRepo = new(mocks.ITokenRepository)
	s.mockJWTService = new(mocks.IJWTService)
	s.mockEmailVerifier = new(mocks.IEmailVerifier)
	s.mockEmailSender = new(mocks.IEmailSender)
	s.mockResetRepo = new(mocks.IPasswordResetRepository)

	s.usecase = usecases.NewUserUsecase(
		s.mockUserRepo,
		s.mockPasswordSvc,
		s.mockTokenRepo,
		s.mockJWTService,
		s.mockEmailVerifier,
		s.mockEmailSender,
		s.mockResetRepo,
	)
}

func (s *UserUsecaseTestSuite) TestRegisterFirstUserAsAdmin() {
	user := userpkg.User{
		Username: "admin1",
		Password: "Str0ng!Pass",
		Email:    "admin@example.com",
		Fullname: "Admin Guy",
	}

	s.mockUserRepo.On("ExistsByUsername", mock.Anything, user.Username).Return(false, nil)
	s.mockUserRepo.On("ExistsByEmail", mock.Anything, user.Email).Return(false, nil)
	s.mockUserRepo.On("CountUsers", mock.Anything).Return(int64(0), nil)
	s.mockEmailVerifier.On("IsRealEmail", user.Email).Return(true, nil)
	s.mockPasswordSvc.On("HashPassword", user.Password).Return("hashed-pass", nil)

	createdUser := user
	createdUser.Role = "admin"
	createdUser.Password = ""

	s.mockUserRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("userpkg.User")).Return(createdUser, nil)

	result, err := s.usecase.RegisterUser(s.ctx, user)
	s.Require().NoError(err)
	s.Equal("admin", result.Role)
	s.Equal("", result.Password)
}

func (s *UserUsecaseTestSuite) TestLoginUser_Success() {
	user := userpkg.User{
		ID:       primitive.NewObjectID(),
		Username: "john",
		Email:    "john@example.com",
		Password: "hashed-pass",
		Role:     "user",
	}

	s.mockUserRepo.On("GetUserByLogin", mock.Anything, "john").Return(user, nil)
	s.mockPasswordSvc.On("ComparePassword", user.Password, "Str0ng!Pass").Return(nil)
	s.mockJWTService.On("GenerateToken", user.ID.Hex(), user.Username, user.Role).Return(userpkg.TokenResult{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
	}, nil)
	s.mockTokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("userpkg.Token")).Return(nil)

	res, access, refresh, err := s.usecase.LoginUser(s.ctx, "john", "Str0ng!Pass")
	s.Require().NoError(err)
	s.Equal("john", res.Username)
	s.Equal("access-token", access)
	s.Equal("refresh-token", refresh)
}

func (s *UserUsecaseTestSuite) TestSendResetOTP_EmailNotFound() {
	s.mockUserRepo.On("ExistsByEmail", mock.Anything, "notfound@example.com").Return(false, nil)

	err := s.usecase.SendResetOTP(s.ctx, "notfound@example.com")
	s.Require().Error(err)
	s.Contains(err.Error(), "email not registered")
}