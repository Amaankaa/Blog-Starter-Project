package userpkg

type UserUsecase interface {
	IRegisterUser(user User) (User, error)
}

// User Infrastructure interfaces
type IJWTService interface {
	GenerateToken(userID, username, role string) (string, error)
	ValidateToken(tokenString string) (map[string]interface{}, error)
}

// PasswordService interface defines password operations
type IPasswordService interface {
	HashPassword(password string) (string, error)
	ComparePassword(hashedPassword, password string) error
}