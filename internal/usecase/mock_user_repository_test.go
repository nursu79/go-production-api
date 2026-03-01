package usecase_test

import (
	"context"

	"github.com/google/uuid"
	"github.com/nursu79/go-production-api/internal/domain"
)

// mockUserRepository is a manual mock implementing domain.UserRepository.
type mockUserRepository struct {
	CreateUserFn     func(ctx context.Context, user *domain.User) (*domain.User, error)
	GetUserByEmailFn func(ctx context.Context, email string) (*domain.User, error)
	GetUserByIDFn    func(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetAllUsersFn    func(ctx context.Context) ([]*domain.User, error)
	DeleteUserFn     func(ctx context.Context, id uuid.UUID) error
}

func (m *mockUserRepository) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	if m.CreateUserFn != nil {
		return m.CreateUserFn(ctx, user)
	}
	return nil, nil
}

func (m *mockUserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.GetUserByEmailFn != nil {
		return m.GetUserByEmailFn(ctx, email)
	}
	return nil, nil
}

func (m *mockUserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if m.GetUserByIDFn != nil {
		return m.GetUserByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockUserRepository) GetAllUsers(ctx context.Context) ([]*domain.User, error) {
	if m.GetAllUsersFn != nil {
		return m.GetAllUsersFn(ctx)
	}
	return nil, nil
}

func (m *mockUserRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if m.DeleteUserFn != nil {
		return m.DeleteUserFn(ctx, id)
	}
	return nil
}
