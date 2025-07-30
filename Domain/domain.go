package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)
// User represents a user entity
type User struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username   string             `bson:"username" json:"username"`
	Fullname   string			  `bson:"fullname" json:"fullname"`
	Email      string             `bson:"email" json:"email"`
	Password   string             `bson:"password" json:"password"`
	Role       string             `bson:"role" json:"role"` // e.g. "admin", "user"
	IsVerified bool               `bson:"isVerified" json:"isVerified"`
}

// UserRepository interface defines user data access operations
type UserRepository interface {
	RegisterUser(user User) (User, error)
}

type UserUsecase interface {
	RegisterUser(user User) (User, error)
}

type JWTService interface {
	GenerateToken(userID, username, role string) (string, error)
	ValidateToken(tokenString string) (map[string]interface{}, error)
}

// PasswordService interface defines password operations
type PasswordService interface {
	HashPassword(password string) (string, error)
	ComparePassword(hashedPassword, password string) error
}