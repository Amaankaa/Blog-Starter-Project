package usecases

import (
	"time"
	"context"
	"errors"
	"testing"

	userpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/user"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// -------------------------------------------------------------
// Mock User Repository
// -------------------------------------------------------------

type MockUserRepo struct {
	users []userpkg.User
}

func (m *MockUserRepo) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	for _, u := range m.users {
		if u.Username == username {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockUserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	for _, u := range m.users {
		if u.Email == email {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockUserRepo) CountUsers(ctx context.Context) (int64, error) {
	return int64(len(m.users)), nil
}

func (m *MockUserRepo) CreateUser(ctx context.Context, user userpkg.User) (userpkg.User, error) {
	m.users = append(m.users, user)
	return user, nil
}

func (m *MockUserRepo) GetUserByLogin(ctx context.Context, login string) (userpkg.User, error) {
	for _, u := range m.users {
		if u.Username == login || u.Email == login {
			return u, nil
		}
	}
	return userpkg.User{}, errors.New("user not found")
}

func (m *MockUserRepo) FindByID(ctx context.Context, userID string) (userpkg.User, error) {
	for _, u := range m.users {
		if u.ID.Hex() == userID{
			return u, nil
		}
	}
	return userpkg.User{}, errors.New("user not found")
}

func (m *MockUserRepo) UpdatePasswordByEmail(ctx context.Context, email, newHashedPassword string) error {
	for i, u := range m.users {
		if u.Email == email {
			m.users[i].Password = newHashedPassword
			return nil
		}
	}
	return errors.New("user not found")
}

// -------------------------------------------------------------
// Mock Password Service
// -------------------------------------------------------------

type MockPasswordService struct{}

func (m *MockPasswordService) HashPassword(password string) (string, error) {
	return "hashed-" + password, nil
}

func (m *MockPasswordService) ComparePassword(hashed, plain string) error {
	return nil // Always accept
}

// Failing Password Service
type FailingPasswordService struct{}

func (f *FailingPasswordService) HashPassword(password string) (string, error) {
	return "hashed-" + password, nil
}

func (f *FailingPasswordService) ComparePassword(_, _ string) error {
	return errors.New("password mismatch")
}

// -------------------------------------------------------------
// Mock Email Verifier
// -------------------------------------------------------------

type MockEmailVerifier struct{}

func (m *MockEmailVerifier) IsRealEmail(email string) (bool, error) {
	return true, nil
}

// -------------------------------------------------------------
// Mock Email Sender
// -------------------------------------------------------------

type MockEmailSender struct{}

func (m *MockEmailSender) SendEmail(to, subject, content string) error {
    return nil // always succeed in tests
}


// -------------------------------------------------------------
// Mock Token Repo
// -------------------------------------------------------------

type MockTokenRepo struct {
	tokens []userpkg.Token
}

func (m *MockTokenRepo) StoreToken(ctx context.Context, token userpkg.Token) error {
	m.tokens = append(m.tokens, token)
	return nil
}

func (m *MockTokenRepo) FindByRefreshToken(ctx context.Context, refreshToken string) (userpkg.Token, error) {
	for _, t := range m.tokens {
		if t.RefreshToken == refreshToken {
			return t, nil
		}
	}
	return userpkg.Token{}, errors.New("token not found")
}

func (m *MockTokenRepo) DeleteByRefreshToken(ctx context.Context, refreshToken string) error {
	for i, t := range m.tokens {
		if t.RefreshToken == refreshToken {
			m.tokens = append(m.tokens[:i], m.tokens[i+1:]...)
			return nil
		}
	}
	return nil
}

// -------------------------------------------------------------
// Mock Password Reset Repo
// -------------------------------------------------------------
type MockPasswordResetRepo struct {
    resets map[string]userpkg.PasswordReset
}

func (m *MockPasswordResetRepo) StoreResetRequest(ctx context.Context, reset userpkg.PasswordReset) error {
    if m.resets == nil {
        m.resets = make(map[string]userpkg.PasswordReset)
    }
    m.resets[reset.Email] = reset
    return nil
}

func (m *MockPasswordResetRepo) GetResetRequest(ctx context.Context, email string) (userpkg.PasswordReset, error) {
    if m.resets == nil {
        return userpkg.PasswordReset{}, errors.New("no reset found")
    }
    reset, ok := m.resets[email]
    if !ok {
        return userpkg.PasswordReset{}, errors.New("no reset found")
    }
    return reset, nil
}

func (m *MockPasswordResetRepo) DeleteResetRequest(ctx context.Context, email string) error {
    if m.resets == nil {
        return nil
    }
    delete(m.resets, email)
    return nil
}

func (m *MockPasswordResetRepo) IncrementAttemptCount(ctx context.Context, email string) error {
    if m.resets == nil {
        return errors.New("no reset data")
    }

    reset, ok := m.resets[email]
    if !ok {
        return errors.New("reset not found")
    }

    reset.AttemptCount++
    m.resets[email] = reset // update the stored value

    return nil
}

// -------------------------------------------------------------
// Mock JWT Service
// -------------------------------------------------------------

type MockJWTService struct {
	UserID string
}

func (m *MockJWTService) GenerateToken(userID, username, role string) (userpkg.TokenResult, error) {
	return userpkg.TokenResult{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
	}, nil
}

func (m *MockJWTService) ValidateToken(token string) (map[string]interface{}, error) {
	return map[string]interface{}{"_id": m.UserID}, nil
}

// -------------------------------------------------------------
// Test Suite
// -------------------------------------------------------------

type UserUsecaseTestSuite struct {
	suite.Suite
	mockRepo     *MockUserRepo
	mockPassword *MockPasswordService
	mockEmail    *MockEmailVerifier
	usecase      *UserUsecase
	ctx          context.Context
}

func TestUserUsecaseTestSuite(t *testing.T) {
	suite.Run(t, new(UserUsecaseTestSuite))
}

func (s *UserUsecaseTestSuite) SetupTest() {
	s.mockRepo = &MockUserRepo{}
	s.mockPassword = &MockPasswordService{}
	s.mockEmail = &MockEmailVerifier{}
	mockTokenRepo := &MockTokenRepo{}
	mockJWT := &MockJWTService{}
	mockEmailSender := &MockEmailSender{}
	mockResetRepo := &MockPasswordResetRepo{}

	s.usecase = NewUserUsecase(
		s.mockRepo,
		s.mockPassword,
		mockTokenRepo,
		mockJWT,
		s.mockEmail,
		mockEmailSender,
		mockResetRepo,
	)

	s.ctx = context.Background()
}

// -------------------------------------------------------------
// Register Tests
// -------------------------------------------------------------

func (s *UserUsecaseTestSuite) TestRegisterFirstUserAsAdmin() {
	user := userpkg.User{
		Username: "admin1",
		Password: "Str0ng!Pass",
		Email:    "admin@example.com",
		Fullname: "Admin Guy",
	}
	created, err := s.usecase.RegisterUser(s.ctx, user)
	s.Require().NoError(err)
	s.Equal("admin", created.Role)
	s.Equal("", created.Password)
}

func (s *UserUsecaseTestSuite) TestRegisterSecondUserAsNormal() {
	s.mockRepo.users = append(s.mockRepo.users, userpkg.User{
		Username: "admin1", Email: "admin@example.com",
	})

	user := userpkg.User{
		Username: "john",
		Password: "Str0ng!Pass",
		Email:    "john@example.com",
		Fullname: "John Doe",
	}
	created, err := s.usecase.RegisterUser(s.ctx, user)
	s.Require().NoError(err)
	s.Equal("user", created.Role)
	s.Equal("", created.Password)
}

func (s *UserUsecaseTestSuite) TestRejectsInvalidEmailFormat() {
	user := userpkg.User{
		Username: "user1",
		Password: "Str0ng!Pass",
		Email:    "not-an-email",
		Fullname: "Bad Email",
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
		Fullname: "Weak Pass",
	}
	_, err := s.usecase.RegisterUser(s.ctx, user)
	s.Require().Error(err)
	s.Contains(err.Error(), "password must be")
}

func (s *UserUsecaseTestSuite) TestRejectsDuplicateUsername() {
	s.mockRepo.users = append(s.mockRepo.users, userpkg.User{Username: "taken"})

	user := userpkg.User{
		Username: "taken",
		Password: "Str0ng!Pass",
		Email:    "new@example.com",
		Fullname: "Cloned",
	}
	_, err := s.usecase.RegisterUser(s.ctx, user)
	s.Require().Error(err)
	s.Contains(err.Error(), "username already taken")
}

func (s *UserUsecaseTestSuite) TestRejectsDuplicateEmail() {
	s.mockRepo.users = append(s.mockRepo.users, userpkg.User{Email: "duplicate@example.com"})

	user := userpkg.User{
		Username: "new",
		Password: "Str0ng!Pass",
		Email:    "duplicate@example.com",
		Fullname: "Clone 2",
	}
	_, err := s.usecase.RegisterUser(s.ctx, user)
	s.Require().Error(err)
	s.Contains(err.Error(), "email already taken")
}

func (s *UserUsecaseTestSuite) TestRejectsEmptyFields() {
	user := userpkg.User{}
	_, err := s.usecase.RegisterUser(s.ctx, user)
	s.Require().Error(err)
	s.Contains(err.Error(), "all fields are required")
}

// Optional: test email unreachable or verifier errors
type FailingEmailVerifier struct{}

func (f *FailingEmailVerifier) IsRealEmail(email string) (bool, error) {
	return false, errors.New("SMTP error")
}

func (s *UserUsecaseTestSuite) TestFailsIfEmailVerifierErrors() {
	s.usecase.emailVerifier = &FailingEmailVerifier{}
	user := userpkg.User{
		Username: "failverifier",
		Password: "Str0ng!Pass",
		Email:    "email@example.com",
		Fullname: "Verifier Error",
	}
	_, err := s.usecase.RegisterUser(s.ctx, user)
	s.Require().Error(err)
	s.Contains(err.Error(), "failed to verify email")
}

// -------------------------------------------------------------
// Login Tests
// -------------------------------------------------------------

func (s *UserUsecaseTestSuite) TestLoginSuccess() {
	s.mockRepo.users = append(s.mockRepo.users, userpkg.User{
		Username: "john",
		Email:    "john@example.com",
		Password: "hashed-Str0ng!Pass",
		Role:     "user",
	})

	result, access, refresh, err := s.usecase.LoginUser(s.ctx, "john", "Str0ng!Pass")
	s.Require().NoError(err)
	s.Equal("john", result.Username)
	s.Equal("access-token", access)
	s.Equal("refresh-token", refresh)
}

func (s *UserUsecaseTestSuite) TestLoginWrongPassword() {
	s.usecase.passwordSvc = &FailingPasswordService{}

	s.mockRepo.users = append(s.mockRepo.users, userpkg.User{
		Username: "john",
		Email:    "john@example.com",
		Password: "hashed-Str0ng!Pass",
	})

	_, _, _, err := s.usecase.LoginUser(s.ctx, "john", "wrongpass")
	s.Require().Error(err)
	s.Contains(err.Error(), "invalid credentials")
}

func (s *UserUsecaseTestSuite) TestLoginUserNotFound() {
	_, _, _, err := s.usecase.LoginUser(s.ctx, "ghost", "whatever")
	s.Require().Error(err)
	s.Contains(err.Error(), "invalid credentials")
}

func (s *UserUsecaseTestSuite) TestRefreshTokenSuccess() {
	user := userpkg.User{
		ID:       primitive.NewObjectID(),
		Username: "john",
		Role:     "user",
	}
	s.mockRepo.users = append(s.mockRepo.users, user)

	mockTokenRepo := &MockTokenRepo{
		tokens: []userpkg.Token{
			{
				UserID:       user.ID,
				RefreshToken: "refresh-token",
				ExpiresAt:    time.Now().Add(1 * time.Hour),
			},
		},
	}
	mockJWT := &MockJWTService{UserID: user.ID.Hex()}
	mockEmailSender := &MockEmailSender{}
	mockResetRepo := &MockPasswordResetRepo{}

	s.usecase = NewUserUsecase(
		s.mockRepo,
		s.mockPassword,
		mockTokenRepo,
		mockJWT,
		s.mockEmail,
		mockEmailSender,
		mockResetRepo,
	)

	result, err := s.usecase.RefreshToken(s.ctx, "refresh-token")
	s.Require().NoError(err)
	s.Equal("access-token", result.AccessToken)
	s.Equal("refresh-token", result.RefreshToken)
}

// -------------------------------------------------------------
// Forgot Password Flow Tests
// -------------------------------------------------------------

func (s *UserUsecaseTestSuite) TestSendResetOTP_Success() {
	email := "otpuser@example.com"
	s.mockRepo.users = append(s.mockRepo.users, userpkg.User{
		Email: email,
	})

	err := s.usecase.SendResetOTP(s.ctx, email)
	s.Require().NoError(err)

	stored, err := s.usecase.passwordResetRepo.GetResetRequest(s.ctx, email)
	s.Require().NoError(err)
	s.Equal(email, stored.Email)
	s.Len(stored.OTP, 6)
	s.False(time.Now().After(stored.ExpiresAt)) // Still valid
	s.Equal(0, stored.AttemptCount)
}

func (s *UserUsecaseTestSuite) TestSendResetOTP_EmailNotFound() {
	err := s.usecase.SendResetOTP(s.ctx, "notfound@example.com")
	s.Require().Error(err)
	s.Contains(err.Error(), "email not registered")
}

func (s *UserUsecaseTestSuite) TestVerifyOTP_Success() {
	email := "verify@example.com"
	otp := "123456"

	_ = s.usecase.passwordResetRepo.StoreResetRequest(s.ctx, userpkg.PasswordReset{
		Email:        email,
		OTP:          otp,
		ExpiresAt:    time.Now().Add(10 * time.Minute),
		AttemptCount: 0,
	})

	err := s.usecase.VerifyOTP(s.ctx, email, otp)
	s.Require().NoError(err)

	_, err = s.usecase.passwordResetRepo.GetResetRequest(s.ctx, email)
	s.Require().Error(err) // Should be deleted
}

func (s *UserUsecaseTestSuite) TestVerifyOTP_Expired() {
	email := "expired@example.com"

	_ = s.usecase.passwordResetRepo.StoreResetRequest(s.ctx, userpkg.PasswordReset{
		Email:        email,
		OTP:          "expired",
		ExpiresAt:    time.Now().Add(-1 * time.Minute),
		AttemptCount: 0,
	})

	err := s.usecase.VerifyOTP(s.ctx, email, "expired")
	s.Require().Error(err)
	s.Contains(err.Error(), "OTP expired")
}

func (s *UserUsecaseTestSuite) TestVerifyOTP_MaxAttempts() {
	email := "max@example.com"

	_ = s.usecase.passwordResetRepo.StoreResetRequest(s.ctx, userpkg.PasswordReset{
		Email:        email,
		OTP:          "654321",
		ExpiresAt:    time.Now().Add(10 * time.Minute),
		AttemptCount: 5,
	})

	err := s.usecase.VerifyOTP(s.ctx, email, "wrong")
	s.Require().Error(err)
	s.Contains(err.Error(), "too many invalid attempts")
}

func (s *UserUsecaseTestSuite) TestVerifyOTP_InvalidOTP() {
	email := "wrongotp@example.com"

	_ = s.usecase.passwordResetRepo.StoreResetRequest(s.ctx, userpkg.PasswordReset{
		Email:        email,
		OTP:          "777777",
		ExpiresAt:    time.Now().Add(10 * time.Minute),
		AttemptCount: 2,
	})

	err := s.usecase.VerifyOTP(s.ctx, email, "123456")
	s.Require().Error(err)
	s.Contains(err.Error(), "invalid OTP")

	// Check if AttemptCount incremented
	stored, _ := s.usecase.passwordResetRepo.GetResetRequest(s.ctx, email)
	s.Equal(3, stored.AttemptCount)
}

func (s *UserUsecaseTestSuite) TestResetPassword_Success() {
	email := "resetme@example.com"

	s.mockRepo.users = append(s.mockRepo.users, userpkg.User{
		Email:    email,
		Password: "old",
	})

	err := s.usecase.ResetPassword(s.ctx, email, "newpass123")
	s.Require().NoError(err)

	user, err := s.mockRepo.GetUserByLogin(s.ctx, email)
	s.Require().NoError(err)
	s.Equal("hashed-newpass123", user.Password)
}