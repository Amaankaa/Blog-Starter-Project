package usecases

import (
	"context"
	"errors"
	"testing"

	userpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/user"
	"github.com/stretchr/testify/suite"
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
// -------------------------------------------------------------
// Mock Password Service
// -------------------------------------------------------------

type MockPasswordService struct{}

func (m *MockPasswordService) HashPassword(password string) (string, error) {
	return "hashed-" + password, nil
}

func (m *MockPasswordService) ComparePassword(hashed, plain string) error {
	return nil
}

// -------------------------------------------------------------
// Mock Email Verifier
// -------------------------------------------------------------

type MockEmailVerifier struct{}

func (m *MockEmailVerifier) IsRealEmail(email string) (bool, error) {
	return true, nil
}

// -------------------------------------------------------------
// Test Suite
// -------------------------------------------------------------

type UserUsecaseTestSuite struct {
	suite.Suite
	mockRepo      *MockUserRepo
	mockPassword  *MockPasswordService
	mockEmail     *MockEmailVerifier
	usecase       *UserUsecase
	ctx           context.Context
}

func TestUserUsecaseTestSuite(t *testing.T) {
	suite.Run(t, new(UserUsecaseTestSuite))
}

func (s *UserUsecaseTestSuite) SetupTest() {
	s.mockRepo = &MockUserRepo{}
	s.mockPassword = &MockPasswordService{}
	s.mockEmail = &MockEmailVerifier{}
	s.usecase = NewUserUsecase(s.mockRepo, s.mockPassword, s.mockEmail)
	s.ctx = context.Background()
}

// -------------------------------------------------------------
// Tests
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
	// Preload first user
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
	user := userpkg.User{
		Username: "",
		Password: "",
		Email:    "",
		Fullname: "",
	}
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