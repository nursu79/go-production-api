package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// User represents the core user entity.
// Notice it maps out the JSON tags carefully to avoid exposing PasswordHash.
type User struct {
	ID        uuid.UUID  `json:"id"`
	Email     string     `json:"email"`
	Password  string     `json:"-"` // Never expose via JSON
	Role      string     `json:"role"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// UserRepository defines the contract for user database operations.
type UserRepository interface {
	CreateUser(ctx context.Context, user *User) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, email, role string) (*User, error)
	GetAllUsers(ctx context.Context) ([]*User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

// UserUsecase defines the contract for user business logic.
type UserUsecase interface {
	Register(ctx context.Context, email, password string) (*User, error)
	Login(ctx context.Context, email, password string) (accessToken string, refreshToken string, err error)
	Logout(ctx context.Context, jti string, expTime time.Duration) error
	GetProfile(ctx context.Context, id uuid.UUID) (*User, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, email, role string) (*User, error)
	GetAllUsers(ctx context.Context) ([]*User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}
