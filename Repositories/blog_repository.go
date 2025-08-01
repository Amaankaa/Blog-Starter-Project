package repositories

import (
	"context"

	blogpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/blog"
	"go.mongodb.org/mongo-driver/mongo"
)

type BlogRepository struct {
	collection *mongo.Collection
	ctx        context.Context
}

func NewBlogRepository(collection *mongo.Collection) *BlogRepository {
	return &BlogRepository{
		collection: collection,
		ctx:        context.Background(),
	}
}

func (br *BlogRepository) CreateBlog(blog *blogpkg.Blog) (*blogpkg.Blog, error) {
	_, err := br.collection.InsertOne(br.ctx, blog)
	if err != nil {
		return nil, err
	}

	return blog, nil
}
