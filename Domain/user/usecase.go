package userpkg

import "context"

type IUserUsecase interface {
	RegisterUser(ctx context.Context, user User) (User, error)
	LoginUser(ctx context.Context, login string, password string) (User, string, string, error)
	RefreshToken(ctx context.Context, refreshToken string)  (TokenResult, error)
}

// User Infrastructure interfaces
type IJWTService interface {
	GenerateToken(userID, username, role string) (TokenResult, error)
	ValidateToken(tokenString string) (map[string]interface{}, error)
}

// PasswordService interface defines password operations
type IPasswordService interface {
	HashPassword(password string) (string, error)
	ComparePassword(hashedPassword, password string) error
}