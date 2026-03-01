package usecase

import (
	"context"
	"net/mail"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/nursu79/go-production-api/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type userUsecase struct {
	userRepo         domain.UserRepository
	jwtSecret        string
	jwtRefreshSecret string
}

// NewUserUsecase injects the repository dependency and JWT secrets.
func NewUserUsecase(userRepo domain.UserRepository, jwtSecret, jwtRefreshSecret string) domain.UserUsecase {
	return &userUsecase{
		userRepo:         userRepo,
		jwtSecret:        jwtSecret,
		jwtRefreshSecret: jwtRefreshSecret,
	}
}

// Register validates input, hashes the password, and invokes the repository.
func (u *userUsecase) Register(ctx context.Context, email, password string) (*domain.User, error) {
	// 1. Validation
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, domain.ErrValidation
	}

	if len(password) < 8 {
		return nil, domain.ErrValidation
	}

	// 2. Hash Password (cost 12 as requested)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, err
	}

	// 3. Delegate to Repository
	user := &domain.User{
		Email:    email,
		Password: string(hashedPassword),
		Role:     "user", // Default domain rule
	}

	return u.userRepo.CreateUser(ctx, user)
}

// Login verifies credentials and evaluates a dual token structure.
func (u *userUsecase) Login(ctx context.Context, email, password string) (string, string, error) {
	// 1. Fetch User By Email
	user, err := u.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		// Do not leak existence info; generic validation error equivalent.
		return "", "", domain.ErrValidation
	}

	// 2. Compare Passwords
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", "", domain.ErrValidation
	}

	// 3. Generate Access Token (15m)
	accessClaims := jwt.MapClaims{
		"sub":     user.ID.String(),
		"role":    user.Role,
		"purpose": "access",
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
		"iat":     time.Now().Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(u.jwtSecret))
	if err != nil {
		return "", "", err
	}

	// 4. Generate Refresh Token (7d)
	refreshClaims := jwt.MapClaims{
		"sub":     user.ID.String(),
		"purpose": "refresh",
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(u.jwtRefreshSecret))
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

// GetProfile delegates to the repository securely fetching a user by ID.
func (u *userUsecase) GetProfile(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return u.userRepo.GetUserByID(ctx, id)
}
