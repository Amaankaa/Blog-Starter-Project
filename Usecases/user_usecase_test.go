package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	userpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/user"
	usecases "github.com/Amaankaa/Blog-Starter-Project/Usecases"
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
		nil,
	)
}

func (s *UserUsecaseTestSuite) TestRegisterFirstUserAsAdmin() {
	// Arrange
	testUser := userpkg.User{
		Username: "adminuser",
		Email:    "admin@example.com",
		Password: "AdminPass123!",
		Fullname: "Admin User",
	}

	s.mockUserRepo.On("CountUsers", s.ctx).Return(int64(0), nil)
	s.mockEmailVerifier.On("IsRealEmail", testUser.Email).Return(true, nil)
	s.mockUserRepo.On("ExistsByUsername", s.ctx, testUser.Username).Return(false, nil)
	s.mockUserRepo.On("ExistsByEmail", s.ctx, testUser.Email).Return(false, nil)
	s.mockPasswordSvc.On("HashPassword", testUser.Password).Return("hashedpassword", nil)
	s.mockUserRepo.On("CreateUser", s.ctx, mock.Anything).Run(func(args mock.Arguments) {
		userArg := args.Get(1).(userpkg.User)
		s.Equal("admin", userArg.Role)
	}).Return(testUser, nil)

	// Act
	result, err := s.usecase.RegisterUser(s.ctx, testUser)

	// Assert
	s.NoError(err)
	s.Equal(testUser.Username, result.Username)
	s.Equal("admin", result.Role)
	s.Empty(result.Password) // Password should be scrubbed
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestRegisterSecondUserAsNormal() {
	// Arrange
	testUser := userpkg.User{
		Username: "normaluser",
		Email:    "user@example.com",
		Password: "UserPass123!",
		Fullname: "Normal User",
	}

	s.mockUserRepo.On("CountUsers", s.ctx).Return(int64(1), nil)
	s.mockEmailVerifier.On("IsRealEmail", testUser.Email).Return(true, nil)
	s.mockUserRepo.On("ExistsByUsername", s.ctx, testUser.Username).Return(false, nil)
	s.mockUserRepo.On("ExistsByEmail", s.ctx, testUser.Email).Return(false, nil)
	s.mockPasswordSvc.On("HashPassword", testUser.Password).Return("hashedpassword", nil)
	s.mockUserRepo.On("CreateUser", s.ctx, mock.Anything).Run(func(args mock.Arguments) {
		userArg := args.Get(1).(userpkg.User)
		s.Equal("user", userArg.Role)
	}).Return(testUser, nil)

	// Act
	result, err := s.usecase.RegisterUser(s.ctx, testUser)

	// Assert
	s.NoError(err)
	s.Equal(testUser.Username, result.Username)
	s.Equal("user", result.Role)
	s.Empty(result.Password)
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestRejectsInvalidEmailFormat() {
	// Arrange
	testUser := userpkg.User{
		Username: "testuser",
		Email:    "invalid-email",
		Password: "ValidPass123!",
		Fullname: "Test User",
	}

	// Act
	_, err := s.usecase.RegisterUser(s.ctx, testUser)

	// Assert
	s.Error(err)
	s.Equal("invalid email format", err.Error())
}

func (s *UserUsecaseTestSuite) TestRejectWeakPassword() {
	// Arrange
	testUser := userpkg.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "weak",
		Fullname: "Test User",
	}
	s.mockEmailVerifier.On("IsRealEmail", mock.Anything).Return(true, nil)

	// Act
	_, err := s.usecase.RegisterUser(s.ctx, testUser)

	// Assert
	s.Error(err)
	s.Equal("password must be at least 8 chars, with upper, lower, number, and special char", err.Error())
}

func (s *UserUsecaseTestSuite) TestRejectDuplicateUsername() {
	// Arrange
	testUser := userpkg.User{
		Username: "existinguser",
		Email:    "test@example.com",
		Password: "ValidPass123!",
		Fullname: "Test User",
	}

	s.mockUserRepo.On("ExistsByUsername", s.ctx, testUser.Username).Return(true, nil)
	s.mockEmailVerifier.On("IsRealEmail", mock.Anything).Return(true, nil)

	// Act
	_, err := s.usecase.RegisterUser(s.ctx, testUser)

	// Assert
	s.Error(err)
	s.Equal("username already taken", err.Error())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestRejectDuplicateEmail() {
	// Arrange
	testUser := userpkg.User{
		Username: "testuser",
		Email:    "existing@example.com",
		Password: "ValidPass123!",
		Fullname: "Test User",
	}

	s.mockUserRepo.On("ExistsByUsername", s.ctx, testUser.Username).Return(false, nil)
	s.mockUserRepo.On("ExistsByEmail", s.ctx, testUser.Email).Return(true, nil)
	s.mockEmailVerifier.On("IsRealEmail", mock.Anything).Return(true, nil)

	// Act
	_, err := s.usecase.RegisterUser(s.ctx, testUser)

	// Assert
	s.Error(err)
	s.Equal("email already taken", err.Error())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestRejectEmptyFields() {
	testCases := []struct {
		name  string
		user  userpkg.User
		field string
	}{
		{"EmptyUsername", userpkg.User{Email: "test@example.com", Password: "Pass123!", Fullname: "Test"}, "username"},
		{"EmptyEmail", userpkg.User{Username: "testuser", Password: "Pass123!", Fullname: "Test"}, "email"},
		{"EmptyPassword", userpkg.User{Username: "testuser", Email: "test@example.com", Fullname: "Test"}, "password"},
		{"EmptyFullname", userpkg.User{Username: "testuser", Email: "test@example.com", Password: "Pass123!"}, "fullname"},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Act
			_, err := s.usecase.RegisterUser(s.ctx, tc.user)

			// Assert
			s.Error(err)
			s.Equal("all fields are required", err.Error())
		})
	}
}

func (s *UserUsecaseTestSuite) TestFailsIfEmailVerifierErrors() {
	// Arrange
	testUser := userpkg.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "ValidPass123!",
		Fullname: "Test User",
	}

	s.mockUserRepo.On("ExistsByUsername", s.ctx, testUser.Username).Return(false, nil)
	s.mockUserRepo.On("ExistsByEmail", s.ctx, testUser.Email).Return(false, nil)
	s.mockEmailVerifier.On("IsRealEmail", testUser.Email).Return(false, errors.New("verification service down"))

	// Act
	_, err := s.usecase.RegisterUser(s.ctx, testUser)

	// Assert
	s.Error(err)
	s.Equal("failed to verify email: verification service down", err.Error())
	s.mockEmailVerifier.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestLoginUser_Success() {
	// Arrange
	login := "testuser"
	password := "ValidPass123!"
	hashedPassword := "hashedpassword"
	userID := primitive.NewObjectID()

	testUser := userpkg.User{
		ID:       userID,
		Username: login,
		Password: hashedPassword,
		Role:     "user",
	}

	tokenRes := userpkg.TokenResult{
		AccessToken:      "access_token",
		RefreshToken:     "refresh_token",
		RefreshExpiresAt: time.Now().Add(24 * time.Hour),
	}

	s.mockUserRepo.On("GetUserByLogin", s.ctx, login).Return(testUser, nil)
	s.mockPasswordSvc.On("ComparePassword", hashedPassword, password).Return(nil)
	s.mockJWTService.On("GenerateToken", userID.Hex(), login, "user").Return(tokenRes, nil)
	s.mockTokenRepo.On("StoreToken", s.ctx, mock.Anything).Return(nil)

	// Act
	user, accessToken, refreshToken, err := s.usecase.LoginUser(s.ctx, login, password)

	// Assert
	s.NoError(err)
	s.Equal(testUser.Username, user.Username)
	s.Equal(tokenRes.AccessToken, accessToken)
	s.Equal(tokenRes.RefreshToken, refreshToken)
	s.Empty(user.Password)
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockPasswordSvc.AssertExpectations(s.T())
	s.mockJWTService.AssertExpectations(s.T())
	s.mockTokenRepo.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestLoginUser_NotFound() {
	// Arrange
	login := "nonexistent"
	password := "password"

	s.mockUserRepo.On("GetUserByLogin", s.ctx, login).Return(userpkg.User{}, errors.New("not found"))

	// Act
	_, _, _, err := s.usecase.LoginUser(s.ctx, login, password)

	// Assert
	s.Error(err)
	s.Equal("invalid credentials", err.Error())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestLoginUser_WrongPassword() {
	// Arrange
	login := "testuser"
	password := "wrongpassword"
	hashedPassword := "hashedpassword"

	testUser := userpkg.User{
		Username: login,
		Password: hashedPassword,
	}

	s.mockUserRepo.On("GetUserByLogin", s.ctx, login).Return(testUser, nil)
	s.mockPasswordSvc.On("ComparePassword", hashedPassword, password).Return(errors.New("mismatch"))

	// Act
	_, _, _, err := s.usecase.LoginUser(s.ctx, login, password)

	// Assert
	s.Error(err)
	s.Equal("invalid credentials", err.Error())
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockPasswordSvc.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestSendResetOTP_EmailNotFound() {
	// Arrange
	email := "nonexistent@example.com"

	s.mockUserRepo.On("ExistsByEmail", s.ctx, email).Return(false, nil)

	// Act
	err := s.usecase.SendResetOTP(s.ctx, email)

	// Assert
	s.Error(err)
	s.Equal("email not registered", err.Error())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestSendResetOTP_Success() {
	// Arrange
	email := "user@example.com"
	hashedOTP := "hashedOTP"

	s.mockUserRepo.On("ExistsByEmail", s.ctx, email).Return(true, nil)
	s.mockEmailSender.On("SendEmail", "user@example.com", "Your OTP Code", mock.Anything).Return(nil)
	s.mockPasswordSvc.
		On("HashPassword", mock.Anything).
		Return(hashedOTP, nil)

	s.mockResetRepo.On("StoreResetRequest", s.ctx, mock.Anything).Return(nil)

	// Act
	err := s.usecase.SendResetOTP(s.ctx, email)

	// Assert
	s.NoError(err)
	s.mockUserRepo.AssertExpectations(s.T())
	s.mockEmailSender.AssertExpectations(s.T())
	s.mockPasswordSvc.AssertExpectations(s.T())
	s.mockResetRepo.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestVerifyOTP_Success() {
	// Arrange
	email := "user@example.com"
	otp := "123456"
	hashedOTP := "hashedOTP"

	storedReset := userpkg.PasswordReset{
		Email:     email,
		OTP:       hashedOTP,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	s.mockResetRepo.On("GetResetRequest", s.ctx, email).Return(storedReset, nil)
	s.mockPasswordSvc.On("ComparePassword", hashedOTP, otp).Return(nil)
	s.mockResetRepo.On("DeleteResetRequest", s.ctx, email).Return(nil)

	// Act
	err := s.usecase.VerifyOTP(s.ctx, email, otp)

	// Assert
	s.NoError(err)
	s.mockResetRepo.AssertExpectations(s.T())
	s.mockPasswordSvc.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestVerifyOTP_Expired() {
	// Arrange
	email := "user@example.com"
	otp := "123456"

	storedReset := userpkg.PasswordReset{
		Email:     email,
		OTP:       "hashedOTP",
		ExpiresAt: time.Now().Add(-10 * time.Minute), // Already expired
	}

	s.mockResetRepo.On("GetResetRequest", s.ctx, email).Return(storedReset, nil)
	s.mockResetRepo.On("DeleteResetRequest", s.ctx, email).Return(nil)

	// Act
	err := s.usecase.VerifyOTP(s.ctx, email, otp)

	// Assert
	s.Error(err)
	s.Equal("OTP expired", err.Error())
	s.mockResetRepo.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestVerifyOTP_MaxAttempts() {
	// Arrange
	email := "user@example.com"
	otp := "123456"

	storedReset := userpkg.PasswordReset{
		Email:        email,
		OTP:          "hashedOTP",
		ExpiresAt:    time.Now().Add(10 * time.Minute),
		AttemptCount: 5, // Max attempts
	}

	s.mockResetRepo.On("GetResetRequest", s.ctx, email).Return(storedReset, nil)
	s.mockResetRepo.On("DeleteResetRequest", s.ctx, email).Return(nil)

	// Act
	err := s.usecase.VerifyOTP(s.ctx, email, otp)

	// Assert
	s.Error(err)
	s.Equal("too many invalid attempts — OTP expired", err.Error())
	s.mockResetRepo.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestVerifyOTP_InvalidOTP() {
	// Arrange
	email := "user@example.com"
	otp := "wrong123"
	hashedOTP := "hashedOTP"

	storedReset := userpkg.PasswordReset{
		Email:     email,
		OTP:       hashedOTP,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	s.mockResetRepo.On("GetResetRequest", s.ctx, email).Return(storedReset, nil)
	s.mockPasswordSvc.On("ComparePassword", hashedOTP, otp).Return(errors.New("mismatch"))
	s.mockResetRepo.On("IncrementAttemptCount", s.ctx, email).Return(nil)

	// Act
	err := s.usecase.VerifyOTP(s.ctx, email, otp)

	// Assert
	s.Error(err)
	s.Equal("invalid OTP", err.Error())
	s.mockResetRepo.AssertExpectations(s.T())
	s.mockPasswordSvc.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestResetPassword_Success() {
	// Arrange
	email := "user@example.com"
	newPassword := "NewPass123!"
	hashedPassword := "hashedNewPassword"

	s.mockPasswordSvc.On("HashPassword", newPassword).Return(hashedPassword, nil)
	s.mockUserRepo.On("UpdatePasswordByEmail", s.ctx, email, hashedPassword).Return(nil)

	// Act
	err := s.usecase.ResetPassword(s.ctx, email, newPassword)

	// Assert
	s.NoError(err)
	s.mockPasswordSvc.AssertExpectations(s.T())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestRefreshToken_Success() {
	// Arrange
	refreshToken := "valid_refresh_token"
	userID := primitive.NewObjectID()
	username := "testuser"
	role := "user"

	claims := map[string]interface{}{
		"_id":      userID.Hex(),
		"username": username,
		"role":     role,
	}

	storedToken := userpkg.Token{
		UserID:       userID,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}

	user := userpkg.User{
		ID:       userID,
		Username: username,
		Role:     role,
	}

	newTokens := userpkg.TokenResult{
		AccessToken:      "new_access_token",
		RefreshToken:     "new_refresh_token",
		RefreshExpiresAt: time.Now().Add(24 * time.Hour),
	}

	s.mockJWTService.On("ValidateToken", refreshToken).Return(claims, nil)
	s.mockTokenRepo.On("FindByRefreshToken", s.ctx, refreshToken).Return(storedToken, nil)
	s.mockUserRepo.On("FindByID", s.ctx, userID.Hex()).Return(user, nil)
	s.mockJWTService.On("GenerateToken", userID.Hex(), username, role).Return(newTokens, nil)
	s.mockTokenRepo.On("DeleteByRefreshToken", s.ctx, refreshToken).Return(nil)
	s.mockTokenRepo.On("StoreToken", s.ctx, mock.Anything).Return(nil)

	// Act
	result, err := s.usecase.RefreshToken(s.ctx, refreshToken)

	// Assert
	s.NoError(err)
	s.Equal(newTokens.AccessToken, result.AccessToken)
	s.Equal(newTokens.RefreshToken, result.RefreshToken)
	s.mockJWTService.AssertExpectations(s.T())
	s.mockTokenRepo.AssertExpectations(s.T())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestRefreshToken_InvalidToken() {
	// Arrange
	invalidToken := "invalid_token"

	s.mockJWTService.On("ValidateToken", invalidToken).Return(nil, errors.New("invalid token"))

	// Act
	_, err := s.usecase.RefreshToken(s.ctx, invalidToken)

	// Assert
	s.Error(err)
	s.Equal("invalid or expired refresh token", err.Error())
	s.mockJWTService.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestRefreshToken_TokenExpired() {
	// Arrange
	expiredToken := "expired_token"
	userID := primitive.NewObjectID()

	claims := map[string]interface{}{
		"_id": userID.Hex(),
	}

	storedToken := userpkg.Token{
		UserID:       userID,
		RefreshToken: expiredToken,
		ExpiresAt:    time.Now().Add(-24 * time.Hour), // Already expired
	}

	s.mockJWTService.On("ValidateToken", expiredToken).Return(claims, nil)
	s.mockTokenRepo.On("FindByRefreshToken", s.ctx, expiredToken).Return(storedToken, nil)

	// Act
	_, err := s.usecase.RefreshToken(s.ctx, expiredToken)

	// Assert
	s.Error(err)
	s.Equal("refresh token expired", err.Error())
	s.mockJWTService.AssertExpectations(s.T())
	s.mockTokenRepo.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestRefreshToken_UserNotFound() {
	// Arrange
	refreshToken := "valid_token"
	userID := primitive.NewObjectID()

	claims := map[string]interface{}{
		"_id": userID.Hex(),
	}

	storedToken := userpkg.Token{
		UserID:       userID,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}

	s.mockJWTService.On("ValidateToken", refreshToken).Return(claims, nil)
	s.mockTokenRepo.On("FindByRefreshToken", s.ctx, refreshToken).Return(storedToken, nil)
	s.mockUserRepo.On("FindByID", s.ctx, userID.Hex()).Return(userpkg.User{}, errors.New("not found"))

	// Act
	_, err := s.usecase.RefreshToken(s.ctx, refreshToken)

	// Assert
	s.Error(err)
	s.mockJWTService.AssertExpectations(s.T())
	s.mockTokenRepo.AssertExpectations(s.T())
	s.mockUserRepo.AssertExpectations(s.T())
}

func (s *UserUsecaseTestSuite) TestLogout_Success() {
	// Arrange
	userID := primitive.NewObjectID().Hex()

	s.mockTokenRepo.
		On("DeleteTokensByUserID", mock.Anything, userID). // <-- Fix here
		Return(nil)

	// Act
	err := s.usecase.Logout(s.ctx, userID)

	// Assert
	s.NoError(err)
	s.mockTokenRepo.AssertCalled(s.T(), "DeleteTokensByUserID", mock.Anything, userID) // <-- Fix here
}

func (s *UserUsecaseTestSuite) TestLogout_FailureFromTokenRepo() {
	// Arrange
	userID := primitive.NewObjectID().Hex()
	expectedErr := errors.New("failed to delete tokens")

	s.mockTokenRepo.
		On("DeleteTokensByUserID", mock.Anything, userID). // <-- Fix here
		Return(expectedErr)

	// Act
	err := s.usecase.Logout(s.ctx, userID)

	// Assert
	s.EqualError(err, expectedErr.Error())
	s.mockTokenRepo.AssertCalled(s.T(), "DeleteTokensByUserID", mock.Anything, userID) // <-- Fix here
}

// TestPromoteUser_CallsRepo ensures PromoteUser calls the repository
func (s *UserUsecaseTestSuite) TestPromoteUser_CallsRepo() {
	id := "user123"
	s.mockUserRepo.On("UpdateUserRoleByID", s.ctx, id, "admin").Return(nil)
	err := s.usecase.PromoteUser(s.ctx, id)
	s.NoError(err)
	s.mockUserRepo.AssertCalled(s.T(), "UpdateUserRoleByID", s.ctx, id, "admin")
}

// TestDemoteUser_CallsRepo ensures DemoteUser calls the repository
func (s *UserUsecaseTestSuite) TestDemoteUser_CallsRepo() {
	id := "user456"
	s.mockUserRepo.On("UpdateUserRoleByID", s.ctx, id, "user").Return(nil)
	err := s.usecase.DemoteUser(s.ctx, id)
	s.NoError(err)
	s.mockUserRepo.AssertCalled(s.T(), "UpdateUserRoleByID", s.ctx, id, "user")
}
