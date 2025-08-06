package controllers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"errors"

	"github.com/Amaankaa/Blog-Starter-Project/Delivery/controllers"
	blogpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/blog"
	"github.com/Amaankaa/Blog-Starter-Project/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type BlogControllerSuite struct {
	suite.Suite
	blogUsecase *mocks.IBlogUsecase
	controller  *controllers.BlogController
	router      *gin.Engine
}

func (s *BlogControllerSuite) SetupTest() {
	s.blogUsecase = new(mocks.IBlogUsecase)
	s.controller = controllers.NewBlogController(s.blogUsecase)
	s.router = gin.Default()

	s.router.POST("/blogs", s.controller.CreateBlog)
	s.router.GET("/blogs", s.controller.GetAllBlogs)
	s.router.GET("/blogs/:id", s.controller.GetBlogByID)
	s.router.PUT("/blogs/:id", s.controller.UpdateBlog)
	s.router.DELETE("/blogs/:id", s.controller.DeleteBlog)
	s.router.GET("/blogs/search", s.controller.SearchBlogs)
	s.router.GET("/blogs/filter", s.controller.FilterByTags)
	s.router.POST("/blogs/:id/like", func(c *gin.Context) {
		// Simulate user_id in context
		c.Set("user_id", "user-1")
		s.controller.LikeBlog(c)
	})
}

func (s *BlogControllerSuite) TestCreateBlog_Success() {
	assert := assert.New(s.T())
	newBlog := &blogpkg.Blog{
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

	s.blogUsecase.On("CreateBlog", mock.Anything, mock.AnythingOfType("*blogpkg.Blog")).Return(expectedBlog, nil)

	body, _ := json.Marshal(newBlog)
	req, _ := http.NewRequest("POST", "/blogs", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	s.router.ServeHTTP(res, req)

	assert.Equal(http.StatusCreated, res.Code)
	assert.Contains(res.Body.String(), "generated-id")
}

func (s *BlogControllerSuite) TestGetAllBlogs_Success() {
	assert := assert.New(s.T())
	// prepare result
	now := time.Now()
	expected := blogpkg.PaginationResponse{
		Data: []blogpkg.Blog{
			{ID: "1", Title: "One", Content: "C1", AuthorID: "A1", Tags: []string{"t1"}, CreatedAt: now, UpdatedAt: now},
			{ID: "2", Title: "Two", Content: "C2", AuthorID: "A2", Tags: []string{"t2"}, CreatedAt: now, UpdatedAt: now},
		},
		Total:      2,
		Page:       2,
		Limit:      5,
		TotalPages: 1,
	}
	pagination := blogpkg.PaginationRequest{Page: 2, Limit: 5}
	s.blogUsecase.On("GetAllBlogs", mock.Anything, pagination).Return(expected, nil)

	req, _ := http.NewRequest("GET", "/blogs?page=2&limit=5", nil)
	res := httptest.NewRecorder()
	s.router.ServeHTTP(res, req)

	assert.Equal(http.StatusOK, res.Code)
	var resp blogpkg.PaginationResponse
	err := json.Unmarshal(res.Body.Bytes(), &resp)
	assert.NoError(err)
	assert.Equal(expected.Total, resp.Total)
	assert.Equal(expected.Page, resp.Page)
	assert.Equal(expected.Limit, resp.Limit)
	assert.Equal(expected.TotalPages, resp.TotalPages)
	assert.Len(resp.Data, len(expected.Data))
	for i, exp := range expected.Data {
		got := resp.Data[i]
		assert.Equal(exp.ID, got.ID)
		assert.Equal(exp.Title, got.Title)
		assert.Equal(exp.Content, got.Content)
		assert.Equal(exp.AuthorID, got.AuthorID)
		assert.ElementsMatch(exp.Tags, got.Tags)
		assert.WithinDuration(exp.CreatedAt, got.CreatedAt, time.Second)
		assert.WithinDuration(exp.UpdatedAt, got.UpdatedAt, time.Second)
	}
}

func (s *BlogControllerSuite) TestGetAllBlogs_Error() {
	assert := assert.New(s.T())
	pagination := blogpkg.PaginationRequest{Page: 1, Limit: 10}
	s.blogUsecase.On("GetAllBlogs", mock.Anything, pagination).Return(blogpkg.PaginationResponse{}, errors.New("repo error"))

	req, _ := http.NewRequest("GET", "/blogs", nil)
	res := httptest.NewRecorder()
	s.router.ServeHTTP(res, req)

	assert.Equal(http.StatusInternalServerError, res.Code)
	assert.Contains(res.Body.String(), "repo error")
}

func (s *BlogControllerSuite) TestGetBlogByID_Success() {
	assert := assert.New(s.T())
	id := "blog-1"
	expected := &blogpkg.Blog{ID: id, Title: "Title1", Content: "Content1", AuthorID: "A1", Tags: []string{"tag1"}}
	s.blogUsecase.On("GetBlogByID", mock.Anything, id).Return(expected, nil)

	req, _ := http.NewRequest("GET", "/blogs/"+id, nil)
	res := httptest.NewRecorder()
	s.router.ServeHTTP(res, req)

	assert.Equal(http.StatusOK, res.Code)
	var resp blogpkg.Blog
	err := json.Unmarshal(res.Body.Bytes(), &resp)
	assert.NoError(err)
	assert.Equal(expected.ID, resp.ID)
	assert.Equal(expected.Title, resp.Title)
	assert.Equal(expected.Content, resp.Content)
	assert.Equal(expected.AuthorID, resp.AuthorID)
	assert.ElementsMatch(expected.Tags, resp.Tags)
}

func (s *BlogControllerSuite) TestGetBlogByID_Error() {
	assert := assert.New(s.T())
	id := "unknown"
	s.blogUsecase.On("GetBlogByID", mock.Anything, id).Return((*blogpkg.Blog)(nil), errors.New("not found"))

	req, _ := http.NewRequest("GET", "/blogs/"+id, nil)
	res := httptest.NewRecorder()
	s.router.ServeHTTP(res, req)

	assert.Equal(http.StatusNotFound, res.Code)
	assert.Contains(res.Body.String(), "not found")
}

func (s *BlogControllerSuite) TestUpdateBlog_Success() {
	assert := assert.New(s.T())
	id := "blog-1"
	update := &blogpkg.Blog{Title: "Updated", Content: "Updated content", Tags: []string{"t1", "t2"}}
	expected := &blogpkg.Blog{ID: id, Title: "Updated", Content: "Updated content", AuthorID: "A1", Tags: []string{"t1", "t2"}}
	s.blogUsecase.On("UpdateBlog", mock.Anything, id, mock.AnythingOfType("*blogpkg.Blog")).Return(expected, nil)

	body, _ := json.Marshal(update)
	req, _ := http.NewRequest("PUT", "/blogs/"+id, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	s.router.ServeHTTP(res, req)

	assert.Equal(http.StatusOK, res.Code)
	var resp blogpkg.Blog
	err := json.Unmarshal(res.Body.Bytes(), &resp)
	assert.NoError(err)
	assert.Equal(expected.ID, resp.ID)
	assert.Equal(expected.Title, resp.Title)
	assert.Equal(expected.Content, resp.Content)
	assert.ElementsMatch(expected.Tags, resp.Tags)
}

func (s *BlogControllerSuite) TestUpdateBlog_Error() {
	assert := assert.New(s.T())
	id := "blog-1"
	s.blogUsecase.On("UpdateBlog", mock.Anything, id, mock.AnythingOfType("*blogpkg.Blog")).Return((*blogpkg.Blog)(nil), errors.New("update failed"))

	update := &blogpkg.Blog{Title: "Updated", Content: "Updated content", Tags: []string{"t1", "t2"}}
	body, _ := json.Marshal(update)
	req, _ := http.NewRequest("PUT", "/blogs/"+id, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	s.router.ServeHTTP(res, req)

	assert.Equal(http.StatusBadRequest, res.Code)
	assert.Contains(res.Body.String(), "update failed")
}

func (s *BlogControllerSuite) TestDeleteBlog_Success() {
	assert := assert.New(s.T())
	id := "blog-1"
	s.blogUsecase.On("DeleteBlog", mock.Anything, id).Return(nil)

	req, _ := http.NewRequest("DELETE", "/blogs/"+id, nil)
	res := httptest.NewRecorder()
	s.router.ServeHTTP(res, req)

	assert.Equal(http.StatusNoContent, res.Code)
}

func (s *BlogControllerSuite) TestDeleteBlog_Error() {
	assert := assert.New(s.T())
	id := "blog-1"
	s.blogUsecase.On("DeleteBlog", mock.Anything, id).Return(errors.New("delete failed"))

	req, _ := http.NewRequest("DELETE", "/blogs/"+id, nil)
	res := httptest.NewRecorder()
	s.router.ServeHTTP(res, req)

	assert.Equal(http.StatusBadRequest, res.Code)
	assert.Contains(res.Body.String(), "delete failed")
}

func (s *BlogControllerSuite) TestSearchBlogs_Success() {
	assert := assert.New(s.T())
	now := time.Now()
	expected := blogpkg.PaginationResponse{
		Data: []blogpkg.Blog{
			{ID: "1", Title: "Go Mongo", Content: "Learning Go with MongoDB", AuthorID: "author-1", Tags: []string{"go", "mongo"}, CreatedAt: now, UpdatedAt: now},
			{ID: "2", Title: "Go Testing", Content: "Testing in Go is fun", AuthorID: "author-2", Tags: []string{"go", "test"}, CreatedAt: now, UpdatedAt: now},
		},
		Total:      2,
		Page:       1,
		Limit:      10,
		TotalPages: 1,
	}
	pagination := blogpkg.PaginationRequest{Page: 1, Limit: 10}
	s.blogUsecase.On("SearchBlogs", mock.Anything, "go", pagination).Return(expected, nil)

	req, _ := http.NewRequest("GET", "/blogs/search?q=go&page=1&limit=10", nil)
	res := httptest.NewRecorder()
	s.router.ServeHTTP(res, req)

	assert.Equal(http.StatusOK, res.Code)
	var resp blogpkg.PaginationResponse
	err := json.Unmarshal(res.Body.Bytes(), &resp)
	assert.NoError(err)
	assert.Equal(expected.Total, resp.Total)
	assert.Equal(expected.Page, resp.Page)
	assert.Equal(expected.Limit, resp.Limit)
	assert.Equal(expected.TotalPages, resp.TotalPages)
	assert.Len(resp.Data, len(expected.Data))
	for i, exp := range expected.Data {
		got := resp.Data[i]
		assert.Equal(exp.ID, got.ID)
		assert.Equal(exp.Title, got.Title)
		assert.Equal(exp.Content, got.Content)
		assert.Equal(exp.AuthorID, got.AuthorID)
		assert.ElementsMatch(exp.Tags, got.Tags)
		assert.WithinDuration(exp.CreatedAt, got.CreatedAt, time.Second)
		assert.WithinDuration(exp.UpdatedAt, got.UpdatedAt, time.Second)
	}
}

func (s *BlogControllerSuite) TestSearchBlogs_EmptyQuery() {
	assert := assert.New(s.T())
	req, _ := http.NewRequest("GET", "/blogs/search", nil)
	res := httptest.NewRecorder()
	s.router.ServeHTTP(res, req)
	assert.Equal(http.StatusBadRequest, res.Code)
	assert.Contains(res.Body.String(), "Search query is required")
}

func (s *BlogControllerSuite) TestSearchBlogs_ErrorFromUsecase() {
	assert := assert.New(s.T())
	pagination := blogpkg.PaginationRequest{Page: 1, Limit: 10}
	s.blogUsecase.On("SearchBlogs", mock.Anything, "go", pagination).Return(blogpkg.PaginationResponse{}, errors.New("repo error"))

	req, _ := http.NewRequest("GET", "/blogs/search?q=go&page=1&limit=10", nil)
	res := httptest.NewRecorder()
	s.router.ServeHTTP(res, req)

	assert.Equal(http.StatusInternalServerError, res.Code)
	assert.Contains(res.Body.String(), "repo error")
}

func (s *BlogControllerSuite) TestFilterByTags_Success() {
	assert := assert.New(s.T())
	now := time.Now()
	tags := []string{"go"}
	expected := blogpkg.PaginationResponse{
		Data: []blogpkg.Blog{
			{ID: "1", Title: "Go Mongo", Content: "Learning Go with MongoDB", AuthorID: "author-1", Tags: []string{"go", "mongo"}, CreatedAt: now, UpdatedAt: now},
			{ID: "2", Title: "Go Testing", Content: "Testing in Go is fun", AuthorID: "author-2", Tags: []string{"go", "test"}, CreatedAt: now, UpdatedAt: now},
		},
		Total:      2,
		Page:       1,
		Limit:      10,
		TotalPages: 1,
	}
	pagination := blogpkg.PaginationRequest{Page: 1, Limit: 10}
	s.blogUsecase.On("FilterByTags", mock.Anything, tags, pagination).Return(expected, nil)

	req, _ := http.NewRequest("GET", "/blogs/filter?tags=go&page=1&limit=10", nil)
	res := httptest.NewRecorder()
	s.router.ServeHTTP(res, req)

	assert.Equal(http.StatusOK, res.Code)
	var resp blogpkg.PaginationResponse
	err := json.Unmarshal(res.Body.Bytes(), &resp)
	assert.NoError(err)
	assert.Equal(expected.Total, resp.Total)
	assert.Equal(expected.Page, resp.Page)
	assert.Equal(expected.Limit, resp.Limit)
	assert.Equal(expected.TotalPages, resp.TotalPages)
	assert.Len(resp.Data, len(expected.Data))
	for i, exp := range expected.Data {
		got := resp.Data[i]
		assert.Equal(exp.ID, got.ID)
		assert.Equal(exp.Title, got.Title)
		assert.Equal(exp.Content, got.Content)
		assert.Equal(exp.AuthorID, got.AuthorID)
		assert.ElementsMatch(exp.Tags, got.Tags)
		assert.WithinDuration(exp.CreatedAt, got.CreatedAt, time.Second)
		assert.WithinDuration(exp.UpdatedAt, got.UpdatedAt, time.Second)
	}
}

func (s *BlogControllerSuite) TestFilterByTags_EmptyTags() {
	assert := assert.New(s.T())
	req, _ := http.NewRequest("GET", "/blogs/filter", nil)
	res := httptest.NewRecorder()
	s.router.ServeHTTP(res, req)
	assert.Equal(http.StatusBadRequest, res.Code)
	assert.Contains(res.Body.String(), "At least one tag is required")
}

func (s *BlogControllerSuite) TestFilterByTags_ErrorFromUsecase() {
	assert := assert.New(s.T())
	tags := []string{"go"}
	pagination := blogpkg.PaginationRequest{Page: 1, Limit: 10}
	s.blogUsecase.On("FilterByTags", mock.Anything, tags, pagination).Return(blogpkg.PaginationResponse{}, errors.New("repo error"))

	req, _ := http.NewRequest("GET", "/blogs/filter?tags=go&page=1&limit=10", nil)
	res := httptest.NewRecorder()
	s.router.ServeHTTP(res, req)

	assert.Equal(http.StatusInternalServerError, res.Code)
	assert.Contains(res.Body.String(), "repo error")
}

func (s *BlogControllerSuite) TestLikeBlog_Success() {
	assert := assert.New(s.T())
	blogID := "blog-1"
	s.blogUsecase.On("ToggleLike", mock.Anything, blogID, "user-1").Return(nil)

	req, _ := http.NewRequest("POST", "/blogs/"+blogID+"/like", nil)
	res := httptest.NewRecorder()
	s.router.ServeHTTP(res, req)

	assert.Equal(http.StatusOK, res.Code)
	assert.Contains(res.Body.String(), "Toggled like successfully")
}

func (s *BlogControllerSuite) TestLikeBlog_MissingBlogID() {
	assert := assert.New(s.T())
	req, _ := http.NewRequest("POST", "/blogs//like", nil)
	res := httptest.NewRecorder()
	s.router.ServeHTTP(res, req)
	assert.Equal(http.StatusBadRequest, res.Code)
	assert.Contains(res.Body.String(), "Blog ID is required")
}

func (s *BlogControllerSuite) TestLikeBlog_Unauthenticated() {
	assert := assert.New(s.T())
	// Register route without setting user_id
	s.router.POST("/blogs/:id/like-noauth", s.controller.LikeBlog)
	req, _ := http.NewRequest("POST", "/blogs/blog-1/like-noauth", nil)
	res := httptest.NewRecorder()
	s.router.ServeHTTP(res, req)
	assert.Equal(http.StatusUnauthorized, res.Code)
	assert.Contains(res.Body.String(), "User not authenticated")
}

func (s *BlogControllerSuite) TestLikeBlog_ErrorFromUsecase() {
	assert := assert.New(s.T())
	blogID := "blog-1"
	s.blogUsecase.On("ToggleLike", mock.Anything, blogID, "user-1").Return(errors.New("toggle error"))

	req, _ := http.NewRequest("POST", "/blogs/"+blogID+"/like", nil)
	res := httptest.NewRecorder()
	s.router.ServeHTTP(res, req)

	assert.Equal(http.StatusInternalServerError, res.Code)
	assert.Contains(res.Body.String(), "toggle error")
}

func TestBlogControllerSuite(t *testing.T) {
	suite.Run(t, new(BlogControllerSuite))
}
