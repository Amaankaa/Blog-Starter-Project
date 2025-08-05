package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	blogpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/blog"
	usecases "github.com/Amaankaa/Blog-Starter-Project/Usecases"
	"github.com/Amaankaa/Blog-Starter-Project/mocks"

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

func (s *BlogUsecaseSuite) TestGetAllBlogs_Success() {
	assert := assert.New(s.T())
	// prepare pagination request
	pagination := blogpkg.PaginationRequest{Page: 2, Limit: 5}
	// prepare expected response
	now := time.Now()
	expected := blogpkg.PaginationResponse{
		Data: []blogpkg.Blog{
			{ID: "1", Title: "Blog One", Content: "Content1", AuthorID: "A1", Tags: []string{"tag1"}, CreatedAt: now, UpdatedAt: now},
			{ID: "2", Title: "Blog Two", Content: "Content2", AuthorID: "A2", Tags: []string{"tag2"}, CreatedAt: now, UpdatedAt: now},
		},
		Total:      2,
		Page:       2,
		Limit:      5,
		TotalPages: 1,
	}
	s.blogRepo.On("GetAllBlogs", mock.Anything, pagination).Return(expected, nil).Once()
	ctx := context.Background()
	resp, err := s.blogUC.GetAllBlogs(ctx, pagination)
	assert.NoError(err)
	assert.Equal(expected, resp)
	s.blogRepo.AssertExpectations(s.T())
}

func (s *BlogUsecaseSuite) TestGetAllBlogs_Defaults() {
	assert := assert.New(s.T())
	// when page and size not provided, defaults apply
	input := blogpkg.PaginationRequest{Page: 0, Limit: 0}
	normalized := blogpkg.PaginationRequest{Page: 1, Limit: 10}
	expected := blogpkg.PaginationResponse{Data: []blogpkg.Blog{}, Total: 0, Page: 1, Limit: 10, TotalPages: 0}
	s.blogRepo.On("GetAllBlogs", mock.Anything, normalized).Return(expected, nil).Once()
	resp, err := s.blogUC.GetAllBlogs(context.Background(), input)
	assert.NoError(err)
	assert.Equal(expected, resp)
	s.blogRepo.AssertExpectations(s.T())
}

func (s *BlogUsecaseSuite) TestGetAllBlogs_MaxLimit() {
	assert := assert.New(s.T())
	// when Limit exceeds max, limit to 100
	input := blogpkg.PaginationRequest{Page: 1, Limit: 200}
	normalized := blogpkg.PaginationRequest{Page: 1, Limit: 100}
	expected := blogpkg.PaginationResponse{Data: []blogpkg.Blog{}, Total: 0, Page: 1, Limit: 100, TotalPages: 0}
	s.blogRepo.On("GetAllBlogs", mock.Anything, normalized).Return(expected, nil).Once()
	resp, err := s.blogUC.GetAllBlogs(context.Background(), input)
	assert.NoError(err)
	assert.Equal(expected, resp)
	s.blogRepo.AssertExpectations(s.T())
}

func (s *BlogUsecaseSuite) TestGetAllBlogs_Error() {
	assert := assert.New(s.T())
	// simulate repo error
	pagination := blogpkg.PaginationRequest{Page: 1, Limit: 10}
	s.blogRepo.On("GetAllBlogs", mock.Anything, pagination).Return(blogpkg.PaginationResponse{}, errors.New("repo error")).Once()
	resp, err := s.blogUC.GetAllBlogs(context.Background(), pagination)
	assert.Error(err)
	assert.EqualError(err, "repo error")
	assert.Equal(blogpkg.PaginationResponse{}, resp)
	s.blogRepo.AssertExpectations(s.T())
}

func (s *BlogUsecaseSuite) TestGetBlogByID_Success() {
	assert := assert.New(s.T())
	id := "blog-1"
	expected := &blogpkg.Blog{ID: id, Title: "Title1", Content: "Cont1", AuthorID: "A1", Tags: []string{"t1"}}
	s.blogRepo.On("GetBlogByID", id).Return(expected, nil).Once()
	ctx := context.Background()
	result, err := s.blogUC.GetBlogByID(ctx, id)
	assert.NoError(err)
	assert.Equal(expected, result)
	s.blogRepo.AssertExpectations(s.T())
}

func (s *BlogUsecaseSuite) TestGetBlogByID_ErrorEmptyID() {
	assert := assert.New(s.T())
	result, err := s.blogUC.GetBlogByID(context.Background(), "")
	assert.Error(err)
	assert.Nil(result)
	assert.EqualError(err, "blog ID is required")
}

func (s *BlogUsecaseSuite) TestGetBlogByID_ErrorRepo() {
	assert := assert.New(s.T())
	id := "unknown"
	s.blogRepo.On("GetBlogByID", id).Return(nil, errors.New("not found")).Once()
	result, err := s.blogUC.GetBlogByID(context.Background(), id)
	assert.Error(err)
	assert.Nil(result)
	assert.EqualError(err, "not found")
	s.blogRepo.AssertExpectations(s.T())
}

func (s *BlogUsecaseSuite) TestUpdateBlog_Success() {
	assert := assert.New(s.T())
	ctx := context.WithValue(context.Background(), "user_id", "author-1")
	id := "blog-1"
	oldBlog := &blogpkg.Blog{
		ID: id, Title: "Old Title", Content: "Old Content", AuthorID: "author-1", Tags: []string{"t1"}, CreatedAt: time.Now().Add(-time.Hour), UpdatedAt: time.Now().Add(-time.Hour),
	}
	updated := &blogpkg.Blog{
		Title: "New Title", Content: "New Content", Tags: []string{"t1", "t2"},
	}
	finalBlog := &blogpkg.Blog{
		ID: id, Title: "New Title", Content: "New Content", AuthorID: "author-1", Tags: []string{"t1", "t2"}, CreatedAt: oldBlog.CreatedAt, UpdatedAt: time.Now(),
	}
	s.blogRepo.On("GetBlogByID", id).Return(oldBlog, nil).Once()
	s.blogRepo.On("UpdateBlog", id, mock.AnythingOfType("*blogpkg.Blog")).Return(finalBlog, nil).Once()
	result, err := s.blogUC.UpdateBlog(ctx, id, updated)
	assert.NoError(err)
	assert.Equal(finalBlog.Title, result.Title)
	assert.Equal(finalBlog.Content, result.Content)
	assert.ElementsMatch(finalBlog.Tags, result.Tags)
	s.blogRepo.AssertExpectations(s.T())
}

func (s *BlogUsecaseSuite) TestUpdateBlog_NotFound() {
	assert := assert.New(s.T())
	ctx := context.WithValue(context.Background(), "user_id", "author-1")
	id := "not-exist"
	s.blogRepo.On("GetBlogByID", id).Return(nil, errors.New("not found")).Once()
	updated := &blogpkg.Blog{Title: "T", Content: "C", Tags: []string{"t1"}}
	result, err := s.blogUC.UpdateBlog(ctx, id, updated)
	assert.Error(err)
	assert.Nil(result)
	assert.Contains(err.Error(), "not found")
	s.blogRepo.AssertExpectations(s.T())
}

func (s *BlogUsecaseSuite) TestUpdateBlog_Unauthorized() {
	assert := assert.New(s.T())
	ctx := context.WithValue(context.Background(), "user_id", "other-user")
	id := "blog-1"
	oldBlog := &blogpkg.Blog{ID: id, Title: "T", Content: "C", AuthorID: "author-1", Tags: []string{"t1"}, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	s.blogRepo.On("GetBlogByID", id).Return(oldBlog, nil).Once()
	updated := &blogpkg.Blog{Title: "T2", Content: "C2", Tags: []string{"t2"}}
	result, err := s.blogUC.UpdateBlog(ctx, id, updated)
	assert.Error(err)
	assert.Nil(result)
	assert.Contains(err.Error(), "unauthorized")
	s.blogRepo.AssertExpectations(s.T())
}

func (s *BlogUsecaseSuite) TestDeleteBlog_Success() {
	assert := assert.New(s.T())
	ctx := context.WithValue(context.Background(), "user_id", "author-1")
	id := "blog-1"
	blog := &blogpkg.Blog{ID: id, Title: "T", Content: "C", AuthorID: "author-1", Tags: []string{"t1"}, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	s.blogRepo.On("GetBlogByID", id).Return(blog, nil).Once()
	s.blogRepo.On("DeleteBlog", id).Return(nil).Once()
	err := s.blogUC.DeleteBlog(ctx, id)
	assert.NoError(err)
	s.blogRepo.AssertExpectations(s.T())
}

func (s *BlogUsecaseSuite) TestDeleteBlog_NotFound() {
	assert := assert.New(s.T())
	ctx := context.WithValue(context.Background(), "user_id", "author-1")
	id := "not-exist"
	s.blogRepo.On("GetBlogByID", id).Return(nil, errors.New("not found")).Once()
	err := s.blogUC.DeleteBlog(ctx, id)
	assert.Error(err)
	assert.Contains(err.Error(), "not found")
	s.blogRepo.AssertExpectations(s.T())
}

func (s *BlogUsecaseSuite) TestDeleteBlog_Unauthorized() {
	assert := assert.New(s.T())
	ctx := context.WithValue(context.Background(), "user_id", "other-user")
	id := "blog-1"
	blog := &blogpkg.Blog{ID: id, Title: "T", Content: "C", AuthorID: "author-1", Tags: []string{"t1"}, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	s.blogRepo.On("GetBlogByID", id).Return(blog, nil).Once()
	err := s.blogUC.DeleteBlog(ctx, id)
	assert.Error(err)
	assert.Contains(err.Error(), "unauthorized")
	s.blogRepo.AssertExpectations(s.T())
}
