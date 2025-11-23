package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"eduBase/internal/helpers"
	"eduBase/internal/models"
	"eduBase/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

type SchoolSelfHandler struct {
	svc *services.SchoolService
}

func NewSchoolSelfHandler(svc *services.SchoolService) *SchoolSelfHandler {
	return &SchoolSelfHandler{svc: svc}
}

func (h *SchoolSelfHandler) Routes(r chi.Router) {
	r.Route("/school", func(r chi.Router) {
		r.Get("/me", h.GetMySchool)
		r.Put("/me", h.UpdateMySchool)
	})
}

func (h *SchoolSelfHandler) requireSchool(w http.ResponseWriter, r *http.Request) (int, bool) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	role, ok := claims["role"].(string)
	if !ok || role != "school" {
		helpers.Error(w, http.StatusForbidden, "access denied")
		return 0, false
	}
	uid, ok := claims["user_id"].(float64)
	if !ok {
		helpers.Error(w, http.StatusForbidden, "invalid token")
		return 0, false
	}
	return int(uid), true
}

// GetMySchool godoc
// @Summary Получить данные своей школы
// @Description Только для пользователей с ролью School
// @Tags Schools
// @Produce json
// @Success 200 {object} models.School
// @Failure 403 {object} helpers.ErrorResponse
// @Failure 404 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Security BearerAuth
// @Router /school/me [get]
func (h *SchoolSelfHandler) GetMySchool(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.requireSchool(w, r)
	if !ok {
		return
	}

	ctx := context.Background()
	school, err := h.svc.GetByUserID(ctx, userID)
	if err != nil {
		helpers.Error(w, http.StatusNotFound, "school not found")
		return
	}

	helpers.JSON(w, http.StatusOK, school)
}

// UpdateMySchool godoc
// @Summary Обновить данные своей школы
// @Description Только для пользователей с ролью School
// @Tags Schools
// @Accept json
// @Produce json
// @Param request body models.School true "Поля для обновления"
// @Success 200 {object} map[string]string
// @Failure 400 {object} helpers.ErrorResponse
// @Failure 403 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Security BearerAuth
// @Router /school/me [put]
func (h *SchoolSelfHandler) UpdateMySchool(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.requireSchool(w, r)
	if !ok {
		return
	}

	var req models.School
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		helpers.Error(w, http.StatusBadRequest, "name required")
		return
	}

	ctx := context.Background()
	if err := h.svc.UpdateByUserID(ctx, userID, &req); err != nil {
		helpers.Error(w, http.StatusInternalServerError, "failed to update school")
		return
	}

	helpers.JSON(w, http.StatusOK, map[string]string{"status": "updated"})
}
