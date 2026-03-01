package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nursu79/go-production-api/internal/delivery/http/response"
	"github.com/nursu79/go-production-api/internal/domain"
)

type UserHandler struct {
	userUsecase domain.UserUsecase
}

// NewUserHandler instantiates the User delivery endpoints.
func NewUserHandler(userUsecase domain.UserUsecase) *UserHandler {
	return &UserHandler{
		userUsecase: userUsecase,
	}
}

// RegisterRequest represents the expected payload for registering a user.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterResponse outlines the safe fields to return back strictly filtering password content.
type RegisterResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// Register acts as the POST endpoint for /api/v1/auth/register handling validation wrapping.
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondError(w, domain.ErrValidation)
		return
	}

	// Delegate processing to usecase
	user, err := h.userUsecase.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		response.RespondError(w, err)
		return
	}

	// Map domain response securely into a DTO before sending
	resp := RegisterResponse{
		ID:        user.ID,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}

	response.RespondJSON(w, http.StatusCreated, resp)
}

// LoginRequest represents the user login payload.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse returns the dual token payloads.
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Login verifies credentials and returns JWT tokens.
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondError(w, domain.ErrValidation)
		return
	}

	accessToken, refreshToken, err := h.userUsecase.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		response.RespondError(w, err)
		return
	}

	resp := LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	response.RespondJSON(w, http.StatusOK, resp)
}

// Logout extracts the JTI from the valid token and maps it into Redis to prevent future reuse.
func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Extract JTI safely injected from JWTMiddleware
	jti, ok := r.Context().Value(domain.JTIKey).(string)
	if !ok {
		response.RespondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized context payload missing jti"})
		return
	}

	// We'll set a hardcoded TTL for blacklisting (e.g., matching the access token's max 15m lifetime).
	// In a more complex setup, you'd extract the "exp" claim and calculate the exact remaining seconds.
	err := h.userUsecase.Logout(r.Context(), jti, 15*time.Minute)
	if err != nil {
		response.RespondError(w, err)
		return
	}

	response.RespondJSON(w, http.StatusOK, map[string]string{"message": "logged out successfully"})
}

// GetMe fetches the currently authenticated user's profile making sure to strip out the password.
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	// Extract sub (uuid string) safely injected from JWTMiddleware
	sub, ok := r.Context().Value(domain.UserIDKey).(string)
	if !ok {
		response.RespondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized context payload"})
		return
	}

	id, err := uuid.Parse(sub)
	if err != nil {
		response.RespondJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid user identity format"})
		return
	}

	// Fetch the complete user profile via Usecase
	user, err := h.userUsecase.GetProfile(r.Context(), id)
	if err != nil {
		response.RespondError(w, err)
		return
	}

	// Mask out the password via the clean DTO struct
	resp := RegisterResponse{
		ID:        user.ID,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}

	response.RespondJSON(w, http.StatusOK, resp)
}

// UpdateMeRequest represents allowable payload mutations securely.
type UpdateMeRequest struct {
	Email string `json:"email"`
}

// UpdateMe triggers Usecase profile alterations securely mapping context ID protections robustly enforcing user boundaries natively.
func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	sub, ok := r.Context().Value(domain.UserIDKey).(string)
	if !ok {
		response.RespondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized context payload"})
		return
	}

	id, err := uuid.Parse(sub)
	if err != nil {
		response.RespondJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid user identity format"})
		return
	}

	var req UpdateMeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondError(w, domain.ErrValidation)
		return
	}

	// For security, standard users shouldn't be able to elevate their own role through this endpoint. We'd patch this explicitly targeting "user" or mapping nil-strings if empty natively.
	updatedUser, err := h.userUsecase.UpdateProfile(r.Context(), id, req.Email, "")
	if err != nil {
		response.RespondError(w, err)
		return
	}

	resp := RegisterResponse{
		ID:        updatedUser.ID,
		Email:     updatedUser.Email,
		Role:      updatedUser.Role,
		CreatedAt: updatedUser.CreatedAt,
	}

	response.RespondJSON(w, http.StatusOK, resp)
}
