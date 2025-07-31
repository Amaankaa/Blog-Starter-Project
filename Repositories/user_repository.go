package repositories

import (
	"context"
	"errors"
	"time"

	userpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/user"
	utils "github.com/Amaankaa/Blog-Starter-Project/Domain/utils"
	services "github.com/Amaankaa/Blog-Starter-Project/Domain/services"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	collection      *mongo.Collection
	passwordService userpkg.PasswordService
	emailVerifier   services.EmailVerifier
}

func NewUserRepository(
	collection *mongo.Collection,
	passwordService userpkg.PasswordService,
	emailVerifier services.EmailVerifier,
) *UserRepository {
	return &UserRepository{
		collection:      collection,
		passwordService: passwordService,
		emailVerifier:   emailVerifier,
	}
}

func (ur *UserRepository) RegisterUser(user userpkg.User) (userpkg.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Check if fields are empty
	if user.Username == "" || user.Password == "" || user.Email == "" || user.Fullname == "" {
		return userpkg.User{}, errors.New("fields cannot be empty")
	}

	// Check if provided email is valid format
	if !utils.IsValidEmail(user.Email) {
		return userpkg.User{}, errors.New("invalid email format")
	}

	// ✅ Check if the email is real using the email verifier
	isReal, err := ur.emailVerifier.IsRealEmail(user.Email)
	if err != nil {
		return userpkg.User{}, errors.New("failed to verify email: " + err.Error())
	}
	if !isReal {
		return userpkg.User{}, errors.New("email address is not reachable")
	}

	// Check if provided password is strong
	if !utils.IsStrongPassword(user.Password) {
		return userpkg.User{}, errors.New("password must be at least 8 characters with upper, lower, number, and special char")
	}

	// Check if the username already exists
	var existingUsername userpkg.User
	err = ur.collection.FindOne(ctx, bson.M{"username": user.Username}).Decode(&existingUsername)
	if err == nil {
		return userpkg.User{}, errors.New("username already taken")
	}
	if err != mongo.ErrNoDocuments {
		return userpkg.User{}, err
	}

	// Check if the email already exists
	var existingEmail userpkg.User
	err = ur.collection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&existingEmail)
	if err == nil {
		return userpkg.User{}, errors.New("email already taken")
	}
	if err != mongo.ErrNoDocuments {
		return userpkg.User{}, err
	}

	// Check if this is the first user (make admin if so)
	userCount, err := ur.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return userpkg.User{}, err
	}
	if userCount == 0 {
		user.Role = "admin"
	} else {
		user.Role = "user"
	}

	// Hash the password before storing
	hashedPassword, err := ur.passwordService.IHashPassword(user.Password)
	if err != nil {
		return userpkg.User{}, err
	}
	user.Password = hashedPassword

	user.ID = primitive.NewObjectID()
	user.IsVerified = false

	_, err = ur.collection.InsertOne(ctx, user)
	if err != nil {
		return userpkg.User{}, err
	}

	user.Password = ""
	return user, nil
}