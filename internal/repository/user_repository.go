package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nursu79/go-production-api/internal/domain"
	"github.com/nursu79/go-production-api/internal/repository/storage"
)

type userRepository struct {
	q *storage.Queries
}

// NewUserRepository injects the db pool and initializes the sqlc Queries instance.
func NewUserRepository(db *pgxpool.Pool) domain.UserRepository {
	return &userRepository{
		q: storage.New(db),
	}
}

// CreateUser inserts a user and maps unique constraint violations to ErrDuplicateEmail.
func (r *userRepository) CreateUser(ctx context.Context, req *domain.User) (*domain.User, error) {
	usr, err := r.q.CreateUser(ctx, storage.CreateUserParams{
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
	})

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return nil, domain.ErrDuplicateEmail
		}
		return nil, err
	}

	return toDomainUser(usr), nil
}

// GetUserByEmail fetches a user by email.
func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	usr, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return toDomainUser(usr), nil
}

// GetUserByID fetches a user by UUID.
func (r *userRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	pgID := pgtype.UUID{Bytes: id, Valid: true}
	usr, err := r.q.GetUserByID(ctx, pgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return toDomainUser(usr), nil
}

// toDomainUser maps the SQLC generated model to the abstracted domain User.
func toDomainUser(s storage.User) *domain.User {
	domainUser := &domain.User{
		ID:        s.ID.Bytes,
		Email:     s.Email,
		Password:  s.Password,
		Role:      s.Role,
		CreatedAt: s.CreatedAt.Time,
		UpdatedAt: s.UpdatedAt.Time,
	}

	if s.DeletedAt.Valid {
		domainUser.DeletedAt = &s.DeletedAt.Time
	}

	return domainUser
}
