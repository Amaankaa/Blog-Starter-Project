package repositories_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	userpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/user"
	repositories "github.com/Amaankaa/Blog-Starter-Project/Repositories"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type userRepositoryTestSuite struct {
	suite.Suite
	db         *mongo.Database
	client     *mongo.Client
	ctx        context.Context
	cancel     context.CancelFunc
	collection *mongo.Collection
	repo       *repositories.UserRepository
}

func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(userRepositoryTestSuite))
}

func (s *userRepositoryTestSuite) SetupSuite() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("MONGODB_URI is not set")
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	s.Require().NoError(err)

	s.client = client
	s.db = client.Database("test_blog_db")
	s.collection = s.db.Collection(testUserCollection)
	s.repo = repositories.NewUserRepository(s.collection)

	s.ctx, s.cancel = context.WithTimeout(context.Background(), 10*time.Second)
}

func (s *userRepositoryTestSuite) TearDownSuite() {
	s.collection.Drop(s.ctx)
	s.cancel()
	s.client.Disconnect(s.ctx)
}

func (s *userRepositoryTestSuite) SetupTest() {
	_, err := s.collection.DeleteMany(s.ctx, bson.M{})
	s.Require().NoError(err)
}

func (s *userRepositoryTestSuite) TestCreateUser() {
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
	s.Equal("", created.Password) // password scrubbed
	s.False(created.IsVerified)
}

func (s *userRepositoryTestSuite) TestExistsByUsername() {
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

func (s *userRepositoryTestSuite) TestExistsByEmail() {
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

func (s *userRepositoryTestSuite) TestCountUsers() {
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