package userpkg

import "context"

// IUserRepository defines user data access operations
type IUserRepository interface {
	FindByID(ctx context.Context, userID string) (User, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	CountUsers(ctx context.Context) (int64, error)
	CreateUser(ctx context.Context, user User) (User, error)
	GetUserByLogin(ctx context.Context, login string) (User, error)
}

type ITokenRepository interface {
	StoreToken(ctx context.Context, token Token) error
	FindByRefreshToken(ctx context.Context, refreshToken string) (Token, error)
	DeleteByRefreshToken(ctx context.Context, refreshToken string) error
}