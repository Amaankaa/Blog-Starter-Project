package userpkg


// UserRepository interface defines user data access operations
type UserRepository interface {
	RegisterUser(user User) (User, error)
}
