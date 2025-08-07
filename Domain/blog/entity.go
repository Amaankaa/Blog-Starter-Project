package blogpkg

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Blog struct {
	ID        string    `json:"id" bson:"id"`
	Title     string    `json:"title" bson:"title"`
	Content   string    `json:"content" bson:"content"`
	AuthorID  string    `json:"author_id" bson:"author_id"`
	Tags      []string  `json:"tags" bson:"tags"`
	Likes     []string  `json:"likes" bson:"likes"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
	Views     int       `json:"views" bson:"views"`
}

// PaginationRequest represents pagination parameters
type PaginationRequest struct {
	Page  int `json:"page" form:"page"`
	Limit int `json:"limit" form:"limit"`
}

// PaginationResponse represents paginated response
type PaginationResponse struct {
	Data       []Blog `json:"data"`
	Total      int64  `json:"total"`
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
	TotalPages int    `json:"total_pages"`
}

type Comment struct {
	ID        primitive.ObjectID `json:"id" bson:"id"`
	BlogID    primitive.ObjectID `json:"blog_id" bson:"blog_id"`
	UserID    string             `json:"user_id" bson:"user_id"`
	Content   string             `json:"content" bson:"content"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
}

type AddCommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=1000"`
}