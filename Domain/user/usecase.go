package userpkg

type UserUsecase interface {
	IRegisterUser(user User) (User, error)
}

// User Infrastructure interfaces
type JWTService interface {
	IGenerateToken(userID, username, role string) (string, error)
	IValidateToken(tokenString string) (map[string]interface{}, error)
}

// PasswordService interface defines password operations
type PasswordService interface {
	IHashPassword(password string) (string, error)
	IComparePassword(hashedPassword, password string) error
}