package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/nursu79/go-production-api/internal/delivery/http/response"
	"github.com/nursu79/go-production-api/internal/domain"
)

type AdminHandler struct {
	userUsecase domain.UserUsecase
}

// NewAdminHandler instantiates Admin delivery endpoints sharing core logic.
func NewAdminHandler(userUsecase domain.UserUsecase) *AdminHandler {
	return &AdminHandler{
		userUsecase: userUsecase,
	}
}

// GetAllUsers retrieves the complete DB set of clean User representations stripped from passwords.
func (h *AdminHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userUsecase.GetAllUsers(r.Context())
	if err != nil {
		response.RespondError(w, err)
		return
	}

	// Mask out passwords natively through DTO
	var resp []RegisterResponse
	for _, user := range users {
		resp = append(resp, RegisterResponse{
			ID:        user.ID,
			Email:     user.Email,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
		})
	}

	// Avoid rendering "null" directly in raw JSON parsing slices if blank
	if len(resp) == 0 {
		response.RespondJSON(w, http.StatusOK, []RegisterResponse{})
		return
	}

	response.RespondJSON(w, http.StatusOK, resp)
}

// DeleteUser strips User accounts mapping into valid UUID constructs targeting soft-deletion operations.
func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id format"})
		return
	}

	if err := h.userUsecase.DeleteUser(r.Context(), id); err != nil {
		response.RespondError(w, err)
		return
	}

	// Empty generic response upon successful soft-delete via REST conventions, generally 204 No Content
	w.WriteHeader(http.StatusNoContent)
}
