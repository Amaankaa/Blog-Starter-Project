package repositories

import (
	"context"
	"testing"
	"time"

	userpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/user"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	db         *mongo.Database
	collection *mongo.Collection
	repo       *UserRepository
	ctx        context.Context
}

func TestUserRepositoryTestSuite(t *testing.T) {
	if testMongoClient == nil {
		t.Skip("MongoDB not available")
	}
	suite.Run(t, new(UserRepositoryTestSuite))
}

func (s *UserRepositoryTestSuite) SetupTest() {
	s.ctx = context.Background()
	dbName := "user_repo_test_" + primitive.NewObjectID().Hex()
	s.db = testMongoClient.Database(dbName)
	s.collection = s.db.Collection("users")
	s.repo = NewUserRepository(s.collection)

	// unique index for email & username to prevent duplicates (optional)
	_ = s.collection.Drop(s.ctx) // cleanup collection if needed
}

func (s *UserRepositoryTestSuite) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = s.db.Drop(ctx)
}

func (s *UserRepositoryTestSuite) TestCreateUser() {
	user := userpkg.User{
		Username: "testuser",
		Password: "hashed-pass",
		Email:    "test@example.com",
		Fullname: "Test User",
	}

	created, err := s.repo.CreateUser(s.ctx, user)
	s.Require().NoError(err)
	s.NotEmpty(created.ID)
	s.Equal("testuser", created.Username)
	s.Equal("test@example.com", created.Email)
	s.Equal("", created.Password) // should be scrubbed
	s.False(created.IsVerified)
}

func (s *UserRepositoryTestSuite) TestExistsByUsername() {
	// Insert a user
	_, err := s.repo.CreateUser(s.ctx, userpkg.User{
		Username: "checkuser",
		Password: "secret",
		Email:    "check@example.com",
		Fullname: "Checker",
	})
	s.Require().NoError(err)

	found, err := s.repo.ExistsByUsername(s.ctx, "checkuser")
	s.Require().NoError(err)
	s.True(found)

	notFound, err := s.repo.ExistsByUsername(s.ctx, "ghost")
	s.Require().NoError(err)
	s.False(notFound)
}

func (s *UserRepositoryTestSuite) TestExistsByEmail() {
	_, err := s.repo.CreateUser(s.ctx, userpkg.User{
		Username: "emailer",
		Password: "secret",
		Email:    "mail@example.com",
		Fullname: "Mailer",
	})
	s.Require().NoError(err)

	found, err := s.repo.ExistsByEmail(s.ctx, "mail@example.com")
	s.Require().NoError(err)
	s.True(found)

	notFound, err := s.repo.ExistsByEmail(s.ctx, "notfound@example.com")
	s.Require().NoError(err)
	s.False(notFound)
}

func (s *UserRepositoryTestSuite) TestCountUsers() {
	initial, err := s.repo.CountUsers(s.ctx)
	s.Require().NoError(err)
	s.Equal(int64(0), initial)

	_, err = s.repo.CreateUser(s.ctx, userpkg.User{
		Username: "counter",
		Password: "pw",
		Email:    "count@example.com",
		Fullname: "Counter",
	})
	s.Require().NoError(err)

	count, err := s.repo.CountUsers(s.ctx)
	s.Require().NoError(err)
	s.Equal(int64(1), count)
}