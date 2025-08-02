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
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
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

func (s *blogRepositoryTestSuite) TestGetAllBlogs() {
	assert := assert.New(s.T())

	// Insert multiple blogs
	blogs := []blogpkg.Blog{
		{
			ID:        "id-1",
			Title:     "Blog 1",
			Content:   "Content 1",
			AuthorID:  "author-1",
			Tags:      []string{"go"},
			CreatedAt: time.Now().Add(-3 * time.Hour),
			UpdatedAt: time.Now().Add(-3 * time.Hour),
		},
		{
			ID:        "id-2",
			Title:     "Blog 2",
			Content:   "Content 2",
			AuthorID:  "author-2",
			Tags:      []string{"mongo"},
			CreatedAt: time.Now().Add(-2 * time.Hour),
			UpdatedAt: time.Now().Add(-2 * time.Hour),
		},
		{
			ID:        "id-3",
			Title:     "Blog 3",
			Content:   "Content 3",
			AuthorID:  "author-3",
			Tags:      []string{"test"},
			CreatedAt: time.Now().Add(-1 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * time.Hour),
		},
	}

	for _, blog := range blogs {
		_, err := s.blogRepo.CreateBlog(&blog)
		assert.NoError(err)
	}

	// Test: Get all blogs, page 1, limit 2
	pagination := blogpkg.PaginationRequest{
		Page:     1,
		Limit: 2,
	}
	resp, err := s.blogRepo.GetAllBlogs(s.ctx, pagination)
	assert.NoError(err)
	assert.Equal(int64(3), resp.Total)
	assert.Equal(2, resp.Limit)
	assert.Equal(1, resp.Page)
	assert.Equal(2, len(resp.Data))
	assert.Equal(2, resp.TotalPages)

	// Check order (should be sorted by created_at descending)
	assert.Equal("Blog 3", resp.Data[0].Title)
	assert.Equal("Blog 2", resp.Data[1].Title)

	// Test: Get all blogs, page 2, limit 2
	pagination.Page = 2
	resp2, err := s.blogRepo.GetAllBlogs(s.ctx, pagination)
	assert.NoError(err)
	assert.Equal(1, len(resp2.Data))
	assert.Equal("Blog 1", resp2.Data[0].Title)

	// Test: Get all blogs, limit larger than total
	pagination.Page = 1
	pagination.Limit = 10
	resp3, err := s.blogRepo.GetAllBlogs(s.ctx, pagination)
	assert.NoError(err)
	assert.Equal(3, len(resp3.Data))
	titles := []string{resp3.Data[0].Title, resp3.Data[1].Title, resp3.Data[2].Title}
	assert.ElementsMatch([]string{"Blog 1", "Blog 2", "Blog 3"}, titles)

	// Test: Empty collection
	_, err = s.collection.DeleteMany(s.ctx, bson.M{})
	assert.NoError(err)
	respEmpty, err := s.blogRepo.GetAllBlogs(s.ctx, blogpkg.PaginationRequest{Page: 1, Limit: 2})
	assert.NoError(err)
	assert.Equal(0, len(respEmpty.Data))
	assert.Equal(int64(0), respEmpty.Total)
	assert.Equal(0, respEmpty.TotalPages)
}

func (s *blogRepositoryTestSuite) TestGetBlogByID_Success() {
	assert := assert.New(s.T())
	// Insert a blog
	blog := &blogpkg.Blog{
		ID:        "id-1",
		Title:     "Blog 1",
		Content:   "Content 1",
		AuthorID:  "author-1",
		Tags:      []string{"go"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	created, err := s.blogRepo.CreateBlog(blog)
	assert.NoError(err)
	// Fetch by ID
	fetched, err := s.blogRepo.GetBlogByID(created.ID)
	assert.NoError(err)
	assert.Equal(created.ID, fetched.ID)
	assert.Equal(created.Title, fetched.Title)
	assert.Equal(created.Content, fetched.Content)
	assert.Equal(created.AuthorID, fetched.AuthorID)
}

func (s *blogRepositoryTestSuite) TestGetBlogByID_NotFound() {
	assert := assert.New(s.T())
	_, err := s.blogRepo.GetBlogByID("unknown-id")
	assert.Error(err)
}
