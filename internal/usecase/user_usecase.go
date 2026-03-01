package usecase

import (
	"context"
	"encoding/json"
	"net/mail"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/nursu79/go-production-api/internal/domain"
	"github.com/nursu79/go-production-api/internal/infrastructure/redis"
	"golang.org/x/crypto/bcrypt"
)

type userUsecase struct {
	userRepo         domain.UserRepository
	jwtSecret        string
	jwtRefreshSecret string
	redisClient      *redis.Client
}

// NewUserUsecase injects the repository dependency, Redis infrastructure, and JWT secrets.
func NewUserUsecase(userRepo domain.UserRepository, redisClient *redis.Client, jwtSecret, jwtRefreshSecret string) domain.UserUsecase {
	return &userUsecase{
		userRepo:         userRepo,
		jwtSecret:        jwtSecret,
		jwtRefreshSecret: jwtRefreshSecret,
		redisClient:      redisClient,
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
	accessJTI := uuid.New().String()
	accessClaims := jwt.MapClaims{
		"jti":     accessJTI,
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
	refreshJTI := uuid.New().String()
	refreshClaims := jwt.MapClaims{
		"jti":     refreshJTI,
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

// Logout leverages the internal redis infrastructure mapping token JWT `jti` directly into explicit blacklists avoiding manual DB calls natively gracefully handling offline states cleanly.
func (u *userUsecase) Logout(ctx context.Context, jti string, expTime time.Duration) error {
	if u.redisClient != nil && u.redisClient.Client != nil {
		// Log the JWT inside the cache with its identical expiration payload natively rendering it totally invalid across the infrastructure.
		return u.redisClient.Client.Set(ctx, "blacklist:"+jti, "true", expTime).Err()
	}
	return nil // Graceful degradation explicitly bypassing cache
}

// GetProfile delegates to the repository securely fetching a user by ID utilizing heavy Cache-Aside Redis patterns natively scaling reads up aggressively protecting PG connections actively.
func (u *userUsecase) GetProfile(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	cacheKey := "user:" + id.String()

	// 1. Try Cache
	if u.redisClient != nil && u.redisClient.Client != nil {
		val, err := u.redisClient.Client.Get(ctx, cacheKey).Result()
		if err == nil {
			var user domain.User
			if jsonErr := json.Unmarshal([]byte(val), &user); jsonErr == nil {
				return &user, nil // Cache Hit!
			}
		}
	}

	// 2. Cache Miss -> Fetch from Repository
	user, err := u.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 3. Store in Cache (10 Minutes TTL) natively avoiding blocking DB returns passively inside goroutines.
	if u.redisClient != nil && u.redisClient.Client != nil {
		go func(usr *domain.User, key string) {
			b, err := json.Marshal(usr)
			if err == nil {
				// Use a fresh context for async redis payload insertion natively avoiding dropped TCP cycles mid-flight.
				u.redisClient.Client.Set(context.Background(), key, b, 15*time.Minute)
			}
		}(user, cacheKey)
	}

	return user, nil
}

// UpdateProfile executes data patching protocols ensuring explicit Cache Invalidation occurs strictly rendering stale JSON maps dropped instantly securely.
func (u *userUsecase) UpdateProfile(ctx context.Context, id uuid.UUID, email, role string) (*domain.User, error) {
	// 1. Delegate Updates downward targeting Postgres natively catching Duplicate violations early. 
	updatedUsr, err := u.userRepo.UpdateUser(ctx, id, email, role)
	if err != nil {
		return nil, err
	}

	// 2. Heavy Cache Invalidation! (If PostgreSQL updated, Redis MUST be aggressively purged to avoid Desyncs natively).
	if u.redisClient != nil && u.redisClient.Client != nil {
		cacheKey := "user:" + id.String()
		u.redisClient.Client.Del(context.Background(), cacheKey) // Flush actively mapping background drops mapping non-blocking threads accurately.
	}

	return updatedUsr, nil
}

// GetAllUsers delegates database reads of multiple rows safely.
func (u *userUsecase) GetAllUsers(ctx context.Context) ([]*domain.User, error) {
	return u.userRepo.GetAllUsers(ctx)
}

// DeleteUser safely strips the entity delegating removal directly to soft deletes and clears the cache.
func (u *userUsecase) DeleteUser(ctx context.Context, id uuid.UUID) error {
	// 1. Delete from DB
	err := u.userRepo.DeleteUser(ctx, id)
	if err != nil {
		return err
	}

	// 2. Invalidate cache
	if u.redisClient != nil && u.redisClient.Client != nil {
		cacheKey := "user:" + id.String()
		u.redisClient.Client.Del(context.Background(), cacheKey) // Graceful background invalidation
	}

	return nil
}
