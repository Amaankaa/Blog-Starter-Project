package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	userpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/user"
	"github.com/Amaankaa/Blog-Starter-Project/Usecases"
	"github.com/Amaankaa/Blog-Starter-Project/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserUsecaseTestSuite struct {
	suite.Suite
	ctx               context.Context
	mockUserRepo      *mocks.IUserRepository
	mockPasswordSvc   *mocks.IPasswordService
	mockTokenRepo     *mocks.ITokenRepository
	mockJWTService    *mocks.IJWTService
	mockEmailVerifier *mocks.IEmailVerifier
	mockEmailSender   *mocks.IEmailSender
	mockResetRepo     *mocks.IPasswordResetRepository
	usecase           *usecases.UserUsecase
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

// ---------------------------------------------
// REGISTER TESTS
// ---------------------------------------------

func (s *UserUsecaseTestSuite) TestRegisterFirstUserAsAdmin() {
	user := userpkg.User{
		Username: "admin1",
		Password: "Str0ng!Pass",
		Email:    "admin@example.com",
		Fullname: "Admin Guy",
	}

	s.mockUserRepo.On("ExistsByUsername", s.ctx, user.Username).Return(false, nil)
	s.mockUserRepo.On("ExistsByEmail", s.ctx, user.Email).Return(false, nil)
	s.mockUserRepo.On("CountUsers", s.ctx).Return(int64(0), nil)
	s.mockEmailVerifier.On("IsRealEmail", user.Email).Return(true, nil)
	s.mockPasswordSvc.On("HashPassword", user.Password).Return("hashed-pass", nil)

	expected := user
	expected.Password = ""
	expected.Role = "admin"

	s.mockUserRepo.On("CreateUser", s.ctx, mock.MatchedBy(func(u userpkg.User) bool {
		return u.Username == user.Username && u.Email == user.Email
	})).Return(expected, nil)

	created, err := s.usecase.RegisterUser(s.ctx, user)
	s.Require().NoError(err)
	s.Equal("admin", created.Role)
	s.Equal("", created.Password)
}

func (s *UserUsecaseTestSuite) TestRegisterSecondUserAsNormal() {
	req := userpkg.User{
		Username: "normaluser",
		Email:    "normal@example.com",
		Password: "StrongPass123!",
	}

	// First user already exists (simulate count = 1)
	s.mockUserRepo.On("CountUsers", s.ctx).Return(int64(1), nil)
	s.mockUserRepo.On("ExistsByUsername", s.ctx, req.Username).Return(false, nil)
	s.mockUserRepo.On("ExistsByEmail", s.ctx, req.Email).Return(false, nil)
	s.mockPasswordSvc.On("ComparePassword", req.Password).Return(nil)
	s.mockPasswordSvc.On("HashPassword", req.Password).Return("hashedpassword", nil)

	expectedUser := req
	expectedUser.Password = "hashedpassword"
	expectedUser.Role = "user"

	s.mockUserRepo.On("CreateUser", s.ctx, mock.MatchedBy(func(u userpkg.User) bool {
		return u.Username == expectedUser.Username &&
			u.Email == expectedUser.Email &&
			u.Password == expectedUser.Password &&
			u.Role == expectedUser.Role
	})).Return(nil)

	_, err := s.usecase.RegisterUser(s.ctx, req)

	s.NoError(err)
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockPasswordSvc.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestRejectsInvalidEmailFormat() {
	user := userpkg.User{
		Username: "user1",
		Password: "Str0ng!Pass",
		Email:    "invalid-email",
		Fullname: "Test User",
	}

	_, err := s.usecase.RegisterUser(s.ctx, user)
	s.Require().Error(err)
	s.Contains(err.Error(), "invalid email format")
}

func (s *UserUsecaseTestSuite) TestRejectsWeakPassword() {
	user := userpkg.User{
		Username: "weak",
		Password: "123",
		Email:    "weak@example.com",
		Fullname: "Weak User",
	}

	_, err := s.usecase.RegisterUser(s.ctx, user)
	s.Require().Error(err)
	s.Contains(err.Error(), "password must be")
}

func (s *UserUsecaseTestSuite) TestRejectsDuplicateUsername() {
	user := userpkg.User{
		Username: "taken",
		Password: "Str0ng!Pass",
		Email:    "new@example.com",
		Fullname: "Clone",
	}

	s.mockUserRepo.On("ExistsByUsername", s.ctx, user.Username).Return(true, nil)

	_, err := s.usecase.RegisterUser(s.ctx, user)
	s.Require().Error(err)
	s.Contains(err.Error(), "username already taken")
}

func (s *UserUsecaseTestSuite) TestRejectsDuplicateEmail() {
	user := userpkg.User{
		Username: "new",
		Password: "Str0ng!Pass",
		Email:    "duplicate@example.com",
		Fullname: "Clone 2",
	}

	s.mockUserRepo.On("ExistsByUsername", s.ctx, user.Username).Return(false, nil)
	s.mockUserRepo.On("ExistsByEmail", s.ctx, user.Email).Return(true, nil)

	_, err := s.usecase.RegisterUser(s.ctx, user)
	s.Require().Error(err)
	s.Contains(err.Error(), "email already taken")
}

func (s *UserUsecaseTestSuite) TestRejectsEmptyFields() {
	req := userpkg.User{} // All fields empty

	_, err := s.usecase.RegisterUser(s.ctx, req)

	s.Error(err)
	s.Contains(err.Error(), "all fields are required")
}

func (s *UserUsecaseTestSuite) TestFailsIfEmailVerifierErrors() {
	user := userpkg.User{
		Username: "failverifier",
		Password: "Str0ng!Pass",
		Email:    "fail@example.com",
		Fullname: "Fail User",
	}

	s.mockUserRepo.On("ExistsByUsername", s.ctx, user.Username).Return(false, nil)
	s.mockUserRepo.On("ExistsByEmail", s.ctx, user.Email).Return(false, nil)
	s.mockUserRepo.On("CountUsers", s.ctx).Return(int64(0), nil)
	s.mockEmailVerifier.On("IsRealEmail", user.Email).Return(false, errors.New("SMTP failure"))

	_, err := s.usecase.RegisterUser(s.ctx, user)
	s.Require().Error(err)
	s.Contains(err.Error(), "failed to verify email")
}

// ---------------------------------------------
// LOGIN TESTS
// ---------------------------------------------

func (s *UserUsecaseTestSuite) TestLoginUser_Success() {
	user := userpkg.User{
		ID:       primitive.NewObjectID(),
		Username: "john",
		Email:    "john@example.com",
		Password: "hashed-pass",
		Role:     "user",
	}

	s.mockUserRepo.On("GetUserByLogin", s.ctx, "john").Return(user, nil)
	s.mockPasswordSvc.On("ComparePassword", user.Password, "Str0ng!Pass").Return(nil)
	s.mockJWTService.On("GenerateToken", user.ID.Hex(), user.Username, user.Role).Return(userpkg.TokenResult{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
	}, nil)
	s.mockTokenRepo.On("StoreToken", s.ctx, mock.AnythingOfType("userpkg.Token")).Return(nil)

	result, access, refresh, err := s.usecase.LoginUser(s.ctx, "john", "Str0ng!Pass")
	s.Require().NoError(err)
	s.Equal("john", result.Username)
	s.Equal("access-token", access)
	s.Equal("refresh-token", refresh)
}

func (s *UserUsecaseTestSuite) TestLoginUser_NotFound() {
	s.mockUserRepo.On("GetUserByLogin", s.ctx, "ghost").Return(userpkg.User{}, errors.New("not found"))

	_, _, _, err := s.usecase.LoginUser(s.ctx, "ghost", "password")
	s.Require().Error(err)
	s.Contains(err.Error(), "invalid credentials")
}

func (s *UserUsecaseTestSuite) TestLoginUser_WrongPassword() {
	user := userpkg.User{
		Username: "john",
		Password: "hashed-pass",
	}

	s.mockUserRepo.On("GetUserByLogin", s.ctx, "john").Return(user, nil)
	s.mockPasswordSvc.On("ComparePassword", user.Password, "wrongpass").Return(errors.New("mismatch"))

	_, _, _, err := s.usecase.LoginUser(s.ctx, "john", "wrongpass")
	s.Require().Error(err)
	s.Contains(err.Error(), "invalid credentials")
}

// ---------------------------------------------
// OTP / PASSWORD RESET TESTS
// ---------------------------------------------

func (s *UserUsecaseTestSuite) TestSendResetOTP_EmailNotFound() {
	s.mockUserRepo.On("ExistsByEmail", s.ctx, "notfound@example.com").Return(false, nil)

	err := s.usecase.SendResetOTP(s.ctx, "notfound@example.com")
	s.Require().Error(err)
	s.Contains(err.Error(), "email not registered")
}

func (s *UserUsecaseTestSuite) TestSendResetOTP_Success() {
	email := "otpuser@example.com"

	s.mockUserRepo.On("ExistsByEmail", s.ctx, email).Return(true, nil)
	s.mockResetRepo.On("StoreResetRequest", s.ctx, mock.MatchedBy(func(req userpkg.PasswordReset) bool {
		return req.Email == email && len(req.OTP) == 6
	})).Return(nil)
	s.mockEmailSender.On("SendEmail", email, mock.Anything, mock.Anything).Return(nil)

	err := s.usecase.SendResetOTP(s.ctx, email)
	s.Require().NoError(err)
}

func (s *UserUsecaseTestSuite) TestVerifyOTP_Success() {
	email := "verify@example.com"
	otp := "123456"
	reset := userpkg.PasswordReset{
		Email:        email,
		OTP:          otp,
		ExpiresAt:    time.Now().Add(10 * time.Minute),
		AttemptCount: 0,
	}

	s.mockResetRepo.On("GetResetRequest", s.ctx, email).Return(reset, nil)
	s.mockResetRepo.On("DeleteResetRequest", s.ctx, email).Return(nil)

	err := s.usecase.VerifyOTP(s.ctx, email, otp)
	s.Require().NoError(err)
}

func (s *UserUsecaseTestSuite) TestVerifyOTP_Expired() {
	email := "expired@example.com"
	reset := userpkg.PasswordReset{
		Email:     email,
		OTP:       "000000",
		ExpiresAt: time.Now().Add(-1 * time.Minute),
	}

	s.mockResetRepo.On("GetResetRequest", s.ctx, email).Return(reset, nil)

	err := s.usecase.VerifyOTP(s.ctx, email, "000000")
	s.Require().Error(err)
	s.Contains(err.Error(), "OTP expired")
}

func (s *UserUsecaseTestSuite) TestVerifyOTP_MaxAttempts() {
	email := "max@example.com"
	reset := userpkg.PasswordReset{
		Email:        email,
		OTP:          "123456",
		ExpiresAt:    time.Now().Add(5 * time.Minute),
		AttemptCount: 5,
	}

	s.mockResetRepo.On("GetResetRequest", s.ctx, email).Return(reset, nil)

	err := s.usecase.VerifyOTP(s.ctx, email, "123456")
	s.Require().Error(err)
	s.Contains(err.Error(), "too many invalid attempts")
}

func (s *UserUsecaseTestSuite) TestVerifyOTP_InvalidOTP() {
	email := "wrongotp@example.com"
	reset := userpkg.PasswordReset{
		Email:        email,
		OTP:          "777777",
		ExpiresAt:    time.Now().Add(10 * time.Minute),
		AttemptCount: 2,
	}

	s.mockResetRepo.On("GetResetRequest", s.ctx, email).Return(reset, nil)
	s.mockResetRepo.On("IncrementAttemptCount", s.ctx, email).Return(nil)

	err := s.usecase.VerifyOTP(s.ctx, email, "wrong")
	s.Require().Error(err)
	s.Contains(err.Error(), "invalid OTP")
}

func (s *UserUsecaseTestSuite) TestResetPassword_Success() {
	email := "resetme@example.com"
	newPass := "NewStrong!Pass"
	hashed := "hashed-NewStrong!Pass"

	s.mockPasswordSvc.On("HashPassword", newPass).Return(hashed, nil)
	s.mockUserRepo.On("UpdatePasswordByEmail", s.ctx, email, hashed).Return(nil)

	err := s.usecase.ResetPassword(s.ctx, email, newPass)
	s.Require().NoError(err)
}

func (s *UserUsecaseTestSuite) TestRefreshToken_Success() {
	userID := primitive.NewObjectID()
	token := userpkg.Token{
		UserID:       userID,
		RefreshToken: "refresh-token",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
	}

	s.mockTokenRepo.On("FindByRefreshToken", s.ctx, "refresh-token").Return(token, nil)
	s.mockUserRepo.On("FindByID", s.ctx, userID.Hex()).Return(userpkg.User{
		ID:       userID,
		Username: "john",
		Role:     "user",
	}, nil)
	s.mockJWTService.On("GenerateToken", userID.Hex(), "john", "user").Return(userpkg.TokenResult{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
	}, nil)

	result, err := s.usecase.RefreshToken(s.ctx, "refresh-token")
	s.Require().NoError(err)
	s.Equal("access-token", result.AccessToken)
	s.Equal("refresh-token", result.RefreshToken)
}

func (s *UserUsecaseTestSuite) TestRefreshToken_InvalidToken() {
	s.mockTokenRepo.On("FindByRefreshToken", s.ctx, "invalid-token").
		Return(userpkg.Token{}, errors.New("not found"))

	_, err := s.usecase.RefreshToken(s.ctx, "invalid-token")
	s.Require().Error(err)
	s.Contains(err.Error(), "refresh token invalid")
}

func (s *UserUsecaseTestSuite) TestRefreshToken_TokenExpired() {
	userID := primitive.NewObjectID()
	token := userpkg.Token{
		UserID:       userID,
		RefreshToken: "refresh-token",
		ExpiresAt:    time.Now().Add(-10 * time.Minute), // expired
	}

	s.mockTokenRepo.On("FindByRefreshToken", s.ctx, "refresh-token").Return(token, nil)

	_, err := s.usecase.RefreshToken(s.ctx, "refresh-token")
	s.Require().Error(err)
	s.Contains(err.Error(), "refresh token expired")
}

func (s *UserUsecaseTestSuite) TestRefreshToken_UserNotFound() {
	userID := primitive.NewObjectID()
	token := userpkg.Token{
		UserID:       userID,
		RefreshToken: "refresh-token",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
	}

	s.mockTokenRepo.On("FindByRefreshToken", s.ctx, "refresh-token").Return(token, nil)
	s.mockUserRepo.On("FindByID", s.ctx, userID.Hex()).Return(userpkg.User{}, errors.New("not found"))

	_, err := s.usecase.RefreshToken(s.ctx, "refresh-token")
	s.Require().Error(err)
	s.Contains(err.Error(), "user not found")
}