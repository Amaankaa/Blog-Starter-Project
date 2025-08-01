package repositories

import (
	"errors"
	"context"

	userpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/user"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(collection *mongo.Collection) *UserRepository {
	return &UserRepository{
		collection: collection,
	}
}

// Check if username exists
func (ur *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var user userpkg.User
	err := ur.collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// Check if email exists
func (ur *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var user userpkg.User
	err := ur.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// Count how many users exist (used to decide admin role)
func (ur *UserRepository) CountUsers(ctx context.Context) (int64, error) {
	count, err := ur.collection.CountDocuments(ctx, bson.M{})
	return count, err
}

// Save new user to DB
func (ur *UserRepository) CreateUser(ctx context.Context, user userpkg.User) (userpkg.User, error) {
	user.ID = primitive.NewObjectID()
	user.IsVerified = false

	_, err := ur.collection.InsertOne(ctx, user)
	if err != nil {
		return userpkg.User{}, err
	}

	user.Password = "" // don’t return hashed password
	return user, nil
}

func (ur *UserRepository) GetUserByLogin(ctx context.Context, login string) (userpkg.User, error) {
	var user userpkg.User
	filter := bson.M{
		"$or": []bson.M{
			{"username": login},
			{"email": login},
		},
	}
	err := ur.collection.FindOne(ctx, filter).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return userpkg.User{}, errors.New("user not found")
	}
	return user, err
}

func (ur *UserRepository) FindByID(ctx context.Context, userID string) (userpkg.User, error) {
	var user userpkg.User

	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return userpkg.User{}, err
	}

	err = ur.collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&user)
	if err != nil {
		return userpkg.User{}, err
	}

	return user, nil
}