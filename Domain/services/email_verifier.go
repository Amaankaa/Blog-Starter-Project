package services

type EmailVerifier interface {
    IsRealEmail(email string) (bool, error)
}