package repositories_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	blogpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/blog"
	repositories "github.com/Amaankaa/Blog-Starter-Project/Repositories"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const testBlogCollection = "test_blogs"

type blogRepositoryTestSuite struct {
	suite.Suite
	db         *mongo.Database
	blogRepo   *repositories.BlogRepository
	ctx        context.Context
	cancel     context.CancelFunc
	client     *mongo.Client
	collection *mongo.Collection
}

func TestBlogRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(blogRepositoryTestSuite))
}

func (s *blogRepositoryTestSuite) SetupSuite() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	testMongoURL := os.Getenv("MONGODB_URI")
	if testMongoURL == "" {
		log.Fatal("MONGODB_URI environment variable is not set")
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(testMongoURL))
	s.Require().NoError(err)

	s.client = client
	s.db = client.Database("test_blog_db")
	s.collection = s.db.Collection(testBlogCollection)
	s.blogRepo = repositories.NewBlogRepository(s.collection)
	s.ctx, s.cancel = context.WithTimeout(context.Background(), 10*time.Second)
}

func (s *blogRepositoryTestSuite) TearDownSuite() {
	s.collection.Drop(s.ctx)
	s.cancel()
	s.client.Disconnect(s.ctx)
}

func (s *blogRepositoryTestSuite) SetupTest() {
	// Clean collection before each test
	_, err := s.collection.DeleteMany(s.ctx, bson.M{})
	s.Require().NoError(err)
}

func (s *blogRepositoryTestSuite) TestCreateBlog() {
	assert := assert.New(s.T())

	blog := &blogpkg.Blog{
		ID:        "test-id-123",
		Title:     "Test Blog",
		Content:   "This is a test blog post.",
		AuthorID:  "author-1",
		Tags:      []string{"go", "mongo"},
		CreatedAt: time.Now().Format(time.RFC3339),
		UpdatedAt: time.Now().Format(time.RFC3339),
	}

	created, err := s.blogRepo.CreateBlog(blog)
	assert.NoError(err)
	assert.NotNil(created)
	assert.Equal("test-id-123", created.ID) 
	assert.Equal(blog.Title, created.Title)
	assert.Equal(blog.Content, created.Content)
	assert.Equal(blog.AuthorID, created.AuthorID)

	// Check in DB
	var found blogpkg.Blog
	err = s.collection.FindOne(s.ctx, bson.M{"id": created.ID}).Decode(&found)
	assert.NoError(err)
	assert.Equal(created.Title, found.Title)
}
