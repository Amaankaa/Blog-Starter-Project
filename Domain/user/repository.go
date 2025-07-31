package userpkg


// UserRepository interface defines user data access operations
type IUserRepository interface {
	RegisterUser(user User) (User, error)
}
