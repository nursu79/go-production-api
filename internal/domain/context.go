package domain

type contextKey string

const (
	// UserIDKey is the context key for the authenticated user's ID.
	UserIDKey contextKey = "user_id"

	// UserRoleKey is the context key for the authenticated user's Role.
	UserRoleKey contextKey = "user_role"
)
