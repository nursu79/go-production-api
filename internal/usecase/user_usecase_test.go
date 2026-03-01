package usecase_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nursu79/go-production-api/internal/domain"
	"github.com/nursu79/go-production-api/internal/usecase"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestUserUsecase_Register(t *testing.T) {
	mockRepo := &mockUserRepository{}
	jwtSecret := "test-secret"
	jwtRefresh := "test-refresh"
	userUsecase := usecase.NewUserUsecase(mockRepo, nil, jwtSecret, jwtRefresh)

	t.Run("successful registration", func(t *testing.T) {
		mockRepo.CreateUserFn = func(ctx context.Context, user *domain.User) (*domain.User, error) {
			user.ID = uuid.New()
			return user, nil
		}

		user, err := userUsecase.Register(context.Background(), "test@example.com", "password123")

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, "user", user.Role)

		// Assert password was hashed
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("password123"))
		assert.NoError(t, err)
	})

	t.Run("invalid email format", func(t *testing.T) {
		user, err := userUsecase.Register(context.Background(), "invalid-email", "password123")

		assert.ErrorIs(t, err, domain.ErrValidation)
		assert.Nil(t, user)
	})

	t.Run("password too short", func(t *testing.T) {
		user, err := userUsecase.Register(context.Background(), "test@example.com", "short")

		assert.ErrorIs(t, err, domain.ErrValidation)
		assert.Nil(t, user)
	})

	t.Run("duplicate email", func(t *testing.T) {
		mockRepo.CreateUserFn = func(ctx context.Context, user *domain.User) (*domain.User, error) {
			return nil, domain.ErrDuplicateEmail
		}

		user, err := userUsecase.Register(context.Background(), "duplicate@example.com", "password123")

		assert.ErrorIs(t, err, domain.ErrDuplicateEmail)
		assert.Nil(t, user)
	})
}

func TestUserUsecase_Login(t *testing.T) {
	mockRepo := &mockUserRepository{}
	jwtSecret := "test-secret"
	jwtRefresh := "test-refresh"
	userUsecase := usecase.NewUserUsecase(mockRepo, nil, jwtSecret, jwtRefresh)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)

	t.Run("successful login", func(t *testing.T) {
		mockRepo.GetUserByEmailFn = func(ctx context.Context, email string) (*domain.User, error) {
			return &domain.User{
				ID:       uuid.New(),
				Email:    "test@example.com",
				Password: string(hashedPassword),
				Role:     "user",
			}, nil
		}

		accessToken, refreshToken, err := userUsecase.Login(context.Background(), "test@example.com", "password123")

		assert.NoError(t, err)
		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)
	})

	t.Run("user not found", func(t *testing.T) {
		mockRepo.GetUserByEmailFn = func(ctx context.Context, email string) (*domain.User, error) {
			return nil, domain.ErrNotFound
		}

		accessToken, refreshToken, err := userUsecase.Login(context.Background(), "unknown@example.com", "password123")

		assert.ErrorIs(t, err, domain.ErrValidation) // Should map not found into generic validation effectively
		assert.Empty(t, accessToken)
		assert.Empty(t, refreshToken)
	})

	t.Run("invalid password", func(t *testing.T) {
		mockRepo.GetUserByEmailFn = func(ctx context.Context, email string) (*domain.User, error) {
			return &domain.User{
				ID:       uuid.New(),
				Email:    "test@example.com",
				Password: string(hashedPassword),
				Role:     "user",
			}, nil
		}

		accessToken, refreshToken, err := userUsecase.Login(context.Background(), "test@example.com", "wrongpassword")

		assert.ErrorIs(t, err, domain.ErrValidation)
		assert.Empty(t, accessToken)
		assert.Empty(t, refreshToken)
	})
}
