package repositories

import (
	"context"
	"math"

	blogpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/blog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

// GetBlogByID fetches a blog by its ID
func (br *BlogRepository) GetBlogByID(id string) (*blogpkg.Blog, error) {
	var blog blogpkg.Blog
	// Filter by custom ID field (stored as 'id')
	filter := bson.M{"id": id}
	err := br.collection.FindOne(br.ctx, filter).Decode(&blog)
	if err != nil {
		return nil, err
	}
	return &blog, nil
}

// GetAllBlogs fetches blogs with pagination
func (br *BlogRepository) GetAllBlogs(ctx context.Context, pagination blogpkg.PaginationRequest) (blogpkg.PaginationResponse, error) {
	// Get total count
	total, err := br.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return blogpkg.PaginationResponse{}, err
	}

	// Apply pagination with skip and limit
	offset := int64((pagination.Page - 1) * pagination.Limit)
	findOptions := options.Find().
		SetLimit(int64(pagination.Limit)).
		SetSkip(offset).
		SetSort(bson.D{{Key: "created_at", Value: -1}}) // Sort by created_at descending

	cursor, err := br.collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return blogpkg.PaginationResponse{}, err
	}
	defer cursor.Close(ctx)

	var blogs []blogpkg.Blog
	if err = cursor.All(ctx, &blogs); err != nil {
		return blogpkg.PaginationResponse{}, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(pagination.Limit)))

	return blogpkg.PaginationResponse{
		Data:       blogs,
		Total:      total,
		Page:       pagination.Page,
		Limit:      pagination.Limit,
		TotalPages: totalPages,
	}, nil
}

func (br *BlogRepository) UpdateBlog(id string, blog *blogpkg.Blog) (*blogpkg.Blog, error) {
	filter := bson.M{"id": id}
	update := bson.M{"$set": blog}
	result := br.collection.FindOneAndUpdate(br.ctx, filter, update)
	if result.Err() != nil {
		return nil, result.Err()
	}

	return blog, nil
}

func (br *BlogRepository) DeleteBlog(id string) error {
	filter := bson.M{"id": id}
	_, err := br.collection.DeleteOne(br.ctx, filter)
	if err != nil {
		return err
	}
	return nil
}
