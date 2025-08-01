package userpkg

import "context"

// IUserRepository defines user data access operations
type IUserRepository interface {
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	CountUsers(ctx context.Context) (int64, error)
	CreateUser(ctx context.Context, user User) (User, error)
}