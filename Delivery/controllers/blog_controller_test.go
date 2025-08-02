package controllers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Amaankaa/Blog-Starter-Project/Delivery/controllers"
	"github.com/Amaankaa/Blog-Starter-Project/Domain/blog"
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

func TestBlogControllerSuite(t *testing.T) {
    suite.Run(t, new(BlogControllerSuite))
}