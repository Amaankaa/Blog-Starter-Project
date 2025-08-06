package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/Amaankaa/Blog-Starter-Project/Domain/services"
	userpkg "github.com/Amaankaa/Blog-Starter-Project/Domain/user"
	utils "github.com/Amaankaa/Blog-Starter-Project/Domain/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserUsecase struct {
	userRepo          userpkg.IUserRepository
	passwordSvc       userpkg.IPasswordService
	tokenRepo         userpkg.ITokenRepository
	emailVerifier     services.IEmailVerifier
	emailSender       services.IEmailSender
	jwtService        userpkg.IJWTService
	passwordResetRepo userpkg.IPasswordResetRepository
}

func NewUserUsecase(
	userRepo userpkg.IUserRepository,
	passwordSvc userpkg.IPasswordService,
	tokenRepo userpkg.ITokenRepository,
	jwtService userpkg.IJWTService,
	emailVerifier services.IEmailVerifier,
	emailSender services.IEmailSender,
	passwordResetRepo userpkg.IPasswordResetRepository,
) *UserUsecase {
	return &UserUsecase{
		userRepo:          userRepo,
		passwordSvc:       passwordSvc,
		tokenRepo:         tokenRepo,
		jwtService:        jwtService,
		emailVerifier:     emailVerifier,
		emailSender:       emailSender,
		passwordResetRepo: passwordResetRepo,
	}
}

func (uu *UserUsecase) RegisterUser(ctx context.Context, user userpkg.User) (userpkg.User, error) {
	// Basic field validation
	if user.Username == "" || user.Email == "" || user.Password == "" || user.Fullname == "" {
		return userpkg.User{}, errors.New("all fields are required")
	}

	// Email format
	if !utils.IsValidEmail(user.Email) {
		return userpkg.User{}, errors.New("invalid email format")
	}

	// Real email check
	isReal, err := uu.emailVerifier.IsRealEmail(user.Email)
	if err != nil {
		return userpkg.User{}, errors.New("failed to verify email: " + err.Error())
	}
	if !isReal {
		return userpkg.User{}, errors.New("email is unreachable")
	}

	// Strong password check
	if !utils.IsStrongPassword(user.Password) {
		return userpkg.User{}, errors.New("password must be at least 8 chars, with upper, lower, number, and special char")
	}

	// Username and email uniqueness
	exists, _ := uu.userRepo.ExistsByUsername(ctx, user.Username)
	if exists {
		return userpkg.User{}, errors.New("username already taken")
	}
	exists, _ = uu.userRepo.ExistsByEmail(ctx, user.Email)
	if exists {
		return userpkg.User{}, errors.New("email already taken")
	}

	// Assign role
	count, err := uu.userRepo.CountUsers(ctx)
	if err != nil {
		return userpkg.User{}, err
	}
	if count == 0 {
		user.Role = "admin"
	} else {
		user.Role = "user"
	}

	// Password hashing
	hashed, err := uu.passwordSvc.HashPassword(user.Password)
	if err != nil {
		return userpkg.User{}, err
	}
	user.Password = hashed
	user.ID = primitive.NewObjectID()
	user.IsVerified = false

	_, err = uu.userRepo.CreateUser(ctx, user)
	if err != nil {
		return userpkg.User{}, err
	}

	user.Password = "" // scrub before return
	return user, nil
}

func (uu *UserUsecase) LoginUser(ctx context.Context, login, password string) (userpkg.User, string, string, error) {
	user, err := uu.userRepo.GetUserByLogin(ctx, login)
	if err != nil {
		return userpkg.User{}, "", "", errors.New("invalid credentials")
	}

	if err := uu.passwordSvc.ComparePassword(user.Password, password); err != nil {
		return userpkg.User{}, "", "", errors.New("invalid credentials")
	}

	// Generate tokens
	tokenRes, err := uu.jwtService.GenerateToken(user.ID.Hex(), user.Username, user.Role)
	if err != nil {
		return userpkg.User{}, "", "", err
	}

	// Store tokens
	err = uu.tokenRepo.StoreToken(ctx, userpkg.Token{
		UserID:       user.ID,
		AccessToken:  tokenRes.AccessToken,
		RefreshToken: tokenRes.RefreshToken,
		CreatedAt:    time.Now(),
		ExpiresAt:    tokenRes.RefreshExpiresAt,
	})
	if err != nil {
		return userpkg.User{}, "", "", err
	}

	user.Password = ""
	return user, tokenRes.AccessToken, tokenRes.RefreshToken, nil
}

func (uu *UserUsecase) RefreshToken(ctx context.Context, refreshToken string) (userpkg.TokenResult, error) {
	claims, err := uu.jwtService.ValidateToken(refreshToken)
	if err != nil {
		return userpkg.TokenResult{}, errors.New("invalid or expired refresh token")
	}

	userID, ok := claims["_id"].(string)
	if !ok {
		return userpkg.TokenResult{}, errors.New("invalid token payload")
	}

	// Check if token is stored in DB
	stored, err := uu.tokenRepo.FindByRefreshToken(ctx, refreshToken)
	if err != nil {
		return userpkg.TokenResult{}, errors.New("refresh token not recognized")
	}

	if stored.ExpiresAt.Before(time.Now()) {
		return userpkg.TokenResult{}, errors.New("refresh token expired")
	}

	// Fetch user info (optional, for roles/username)
	user, err := uu.userRepo.FindByID(ctx, userID)
	if err != nil {
		return userpkg.TokenResult{}, err
	}

	// Generate new tokens
	tokens, err := uu.jwtService.GenerateToken(user.ID.Hex(), user.Username, user.Role)
	if err != nil {
		return userpkg.TokenResult{}, err
	}

	// Store new refresh token, remove old
	_ = uu.tokenRepo.DeleteByRefreshToken(ctx, refreshToken)
	_ = uu.tokenRepo.StoreToken(ctx, userpkg.Token{
		UserID:       user.ID,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.RefreshExpiresAt,
		CreatedAt:    time.Now(),
	})

	return tokens, nil
}

func (u *UserUsecase) SendResetOTP(ctx context.Context, email string) error {
	exists, _ := u.userRepo.ExistsByEmail(ctx, email)
	if !exists {
		return errors.New("email not registered")
	}

	otp := utils.GenerateOTP(6)

	err := u.emailSender.SendEmail(email, "Your OTP Code", "Your OTP: "+otp)
	if err != nil {
		return err
	}

	//Hash OTP before storing
	hashedOTP, err := u.passwordSvc.HashPassword(otp)
	if err != nil {
		return err
	}

	reset := userpkg.PasswordReset{
		Email:        email,
		OTP:          hashedOTP,
		ExpiresAt:    time.Now().Add(10 * time.Minute),
		AttemptCount: 0,
	}
	return u.passwordResetRepo.StoreResetRequest(ctx, reset)
}

func (u *UserUsecase) VerifyOTP(ctx context.Context, email, otp string) error {
	stored, err := u.passwordResetRepo.GetResetRequest(ctx, email)
	if err != nil {
		return errors.New("no reset request found")
	}

	if time.Now().After(stored.ExpiresAt) {
		_ = u.passwordResetRepo.DeleteResetRequest(ctx, email)
		return errors.New("OTP expired")
	}

	if stored.AttemptCount >= 5 {
		_ = u.passwordResetRepo.DeleteResetRequest(ctx, email)
		return errors.New("too many invalid attempts — OTP expired")
	}

	if u.passwordSvc.ComparePassword(stored.OTP, otp) != nil {
		// increment attempt count
		_ = u.passwordResetRepo.IncrementAttemptCount(ctx, email)
		return errors.New("invalid OTP")
	}

	// OTP is valid — delete it
	_ = u.passwordResetRepo.DeleteResetRequest(ctx, email)
	return nil
}

func (u *UserUsecase) ResetPassword(ctx context.Context, email, newPassword string) error {
	hashed, err := u.passwordSvc.HashPassword(newPassword)
	if err != nil {
		return err
	}
	return u.userRepo.UpdatePasswordByEmail(ctx, email, hashed)
}

func (u *UserUsecase) Logout(ctx context.Context, userID string) error {
	return u.tokenRepo.DeleteTokensByUserID(ctx, userID)
}

func (uu *UserUsecase) PromoteUser(ctx context.Context, userID string) error {
	return uu.userRepo.UpdateUserRoleByID(ctx, userID, "admin")
}

func (uu *UserUsecase) DemoteUser(ctx context.Context, userID string) error {
	return uu.userRepo.UpdateUserRoleByID(ctx, userID, "user")
}
