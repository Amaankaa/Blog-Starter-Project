package usecases_test

import (
	   "context"
	   "errors"
	   "testing"
	   "time"

	   blogpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/blog"
	   "github.com/Amaankaa/Blog-Starter-Project/mocks"
	   usecases "github.com/Amaankaa/Blog-Starter-Project/Usecases"

	   "github.com/stretchr/testify/assert"
	   "github.com/stretchr/testify/mock"
	   "github.com/stretchr/testify/suite"
)

type BlogUsecaseSuite struct {
	   suite.Suite
	   blogRepo *mocks.IBlogRepository
	   blogUC   *usecases.BlogUsecase
}

func (s *BlogUsecaseSuite) SetupTest() {
	   s.blogRepo = mocks.NewIBlogRepository(s.T())
	   s.blogUC = usecases.NewBlogUsecase(s.blogRepo)
}

func TestBlogUsecaseSuite(t *testing.T) {
	   suite.Run(t, new(BlogUsecaseSuite))
}

func (s *BlogUsecaseSuite) TestCreateBlog_Success() {
	   assert := assert.New(s.T())
	   blog := &blogpkg.Blog{
			   Title:   "Test Blog",
			   Content: "Test content",
			   Tags:    []string{"test", "go"},
	   }
	   expectedBlog := &blogpkg.Blog{
			   ID:        "generated-id",
			   Title:     "Test Blog",
			   Content:   "Test content",
			   AuthorID:  "user123",
			   Tags:      []string{"test", "go"},
			   CreatedAt: time.Now(),
			   UpdatedAt: time.Now(),
	   }
	   s.blogRepo.On("CreateBlog", mock.AnythingOfType("*blogpkg.Blog")).Return(expectedBlog, nil).Once()
	   ctx := context.WithValue(context.Background(), "user_id", "user123")
	   result, err := s.blogUC.CreateBlog(ctx, blog)
	   assert.NoError(err)
	   assert.NotNil(result)
	   assert.Equal(blog.Title, result.Title)
	   assert.Equal(blog.Content, result.Content)
	   assert.Equal("user123", result.AuthorID)
	   assert.NotEmpty(result.ID)
	   assert.NotZero(result.CreatedAt)
	   assert.NotZero(result.UpdatedAt)
	   s.blogRepo.AssertExpectations(s.T())
}

func (s *BlogUsecaseSuite) TestCreateBlog_Error() {
	   assert := assert.New(s.T())
	   blog := &blogpkg.Blog{
			   Title:   "Test Blog",
			   Content: "Test content",
			   Tags:    []string{"test", "go"},
	   }
	   s.blogRepo.On("CreateBlog", mock.AnythingOfType("*blogpkg.Blog")).Return(nil, errors.New("create failed")).Once()
	   ctx := context.WithValue(context.Background(), "user_id", "user123")
	   result, err := s.blogUC.CreateBlog(ctx, blog)
	   assert.Error(err)
	   assert.Nil(result)
	   assert.EqualError(err, "create failed")
	   s.blogRepo.AssertExpectations(s.T())
}

