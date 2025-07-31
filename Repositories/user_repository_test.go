package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/Amaankaa/Blog-Starter-Project/Domain/user"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// -------------------------------------------------------------------
// Mock Password Service
// -------------------------------------------------------------------

type MockPasswordService struct{}

func (m *MockPasswordService) IHashPassword(password string) (string, error) {
	return "hashed-" + password, nil
}

func (f *MockPasswordService) IComparePassword(hashed, plain string) error {
	return nil
}

// -------------------------------------------------------------------
// User Repository Test Suite
// -------------------------------------------------------------------

type UserRepoTestSuite struct {
	suite.Suite
	db    *mongo.Database
	users *mongo.Collection
	repo  *UserRepository
}

func TestUserRepoTestSuite(t *testing.T) {
	if testMongoClient == nil {
		t.Skip("MongoDB client not available")
	}
	suite.Run(t, new(UserRepoTestSuite))
}

func (ts *UserRepoTestSuite) SetupTest() {
	dbName := "blog_test_db_" + primitive.NewObjectID().Hex()
	ts.db = testMongoClient.Database(dbName)
	ts.users = ts.db.Collection("users")

	index := mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err := ts.users.Indexes().CreateOne(context.Background(), index)
	ts.Require().NoError(err)

	ts.repo = NewUserRepository(ts.users, &MockPasswordService{})
}

func (ts *UserRepoTestSuite) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = ts.db.Drop(ctx)
}

// -------------------------------------------------------------------
// Individual Tests
// -------------------------------------------------------------------

func (ts *UserRepoTestSuite) TestAssignsAdminRoleToFirstUser() {
	user := userpkg.User{
		Username: "admin",
		Password: "Str0ng!Pass",
		Email:    "admin@example.com",
		Fullname: "Admin User",
	}

	storedUser, err := ts.repo.RegisterUser(user)
	ts.Require().NoError(err)
	ts.Equal("admin", storedUser.Role)
}

func (ts *UserRepoTestSuite) TestRegistersUserAsNormalUser() {
	// Register admin first
	_, err := ts.repo.RegisterUser(userpkg.User{
		Username: "admin",
		Password: "Str0ng!Pass",
		Email:    "admin@example.com",
		Fullname: "Admin",
	})
	ts.Require().NoError(err)

	// Then register a second user
	newUser := userpkg.User{
		Username: "johndoe",
		Password: "Str0ng!Pass",
		Email:    "john@example.com",
		Fullname: "John Doe",
	}

	storedUser, err := ts.repo.RegisterUser(newUser)
	ts.Require().NoError(err)
	ts.NotEmpty(storedUser.ID)
	ts.Equal("johndoe", storedUser.Username)
	ts.Equal("user", storedUser.Role)
	ts.False(storedUser.IsVerified)
	ts.Empty(storedUser.Password)
}

func (ts *UserRepoTestSuite) TestRejectsWeakPassword() {
	user := userpkg.User{
		Username: "weakman",
		Password: "123",
		Email:    "weak@example.com",
		Fullname: "Weak Man",
	}

	_, err := ts.repo.RegisterUser(user)
	ts.Require().Error(err)
}

func (ts *UserRepoTestSuite) TestRejectsInvalidEmail() {
	user := userpkg.User{
		Username: "bademail",
		Password: "Str0ng!Pass",
		Email:    "not-an-email",
		Fullname: "Bad Email",
	}

	_, err := ts.repo.RegisterUser(user)
	ts.Require().Error(err)
}

func (ts *UserRepoTestSuite) TestRejectsDuplicateEmail() {
	first := userpkg.User{
		Username: "original",
		Password: "Str0ng!Pass",
		Email:    "same@example.com",
		Fullname: "Original User",
	}
	_, err := ts.repo.RegisterUser(first)
	ts.Require().NoError(err)

	dupe := userpkg.User{
		Username: "copycat",
		Password: "Str0ng!Pass",
		Email:    "same@example.com",
		Fullname: "Copy Cat",
	}
	_, err = ts.repo.RegisterUser(dupe)
	ts.Require().Error(err)
}

func (ts *UserRepoTestSuite) TestRejectsDuplicateUsername() {
	first := userpkg.User{
		Username: "reused",
		Password: "Str0ng!Pass",
		Email:    "reused1@example.com",
		Fullname: "First",
	}
	_, err := ts.repo.RegisterUser(first)
	ts.Require().NoError(err)

	dupe := userpkg.User{
		Username: "reused",
		Password: "Str0ng!Pass",
		Email:    "reused2@example.com",
		Fullname: "Second",
	}
	_, err = ts.repo.RegisterUser(dupe)
	ts.Require().Error(err)
}