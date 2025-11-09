package handlers

import (
	"context"
	"eduBase/internal/helpers"
	"eduBase/internal/repository"
	"eduBase/internal/services"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// RooHandler управляет действиями Районного отдела образования (ROO).
type RooHandler struct {
	svc        *services.AuthService
	schoolRepo *repository.SchoolRepository
}

// NewRooHandler конструктор.
func NewRooHandler(authSvc *services.AuthService, schoolRepo *repository.SchoolRepository) *RooHandler {
	return &RooHandler{svc: authSvc, schoolRepo: schoolRepo}
}

// Routes регистрирует роуты ROO.
func (h *RooHandler) Routes(r chi.Router) {
	r.Route("/roo", func(r chi.Router) {
		r.Post("/register_school", h.RegisterSchool)
	})
}

// registerSchoolRequest структура запроса на регистрацию школы.
type registerSchoolRequest struct {
	Email    string `json:"email" example:"school1@example.com"`
	Password string `json:"password" example:"123456"`
	Name     string `json:"name" example:"Школа №1"`
	Director string `json:"director" example:"Иванов Иван"`
}

// RegisterSchool godoc
// @Summary Регистрация школы (ROO)
// @Description Создаёт школу и генерирует пароль автоматически.
// @Tags ROO
// @Accept json
// @Produce json
// @Param input body registerSchoolRequest true "Данные для регистрации"
// @Success 201 {object} map[string]string
// @Failure 400 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Security BearerAuth
// @Router /roo/register-school [post]
func (h *RooHandler) RegisterSchool(w http.ResponseWriter, r *http.Request) {
	var req registerSchoolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	password, err := h.svc.RegisterSchool(
		context.Background(),
		req.Email,
		req.Name,
		req.Director,
		h.schoolRepo,
	)
	if err != nil {
		helpers.Error(w, http.StatusInternalServerError, "failed to register school")
		return
	}

	// Возвращаем сгенерированный пароль, чтобы ROO мог передать школе
	helpers.JSON(w, http.StatusCreated, map[string]string{
		"status":   "school registered",
		"email":    req.Email,
		"password": password,
	})
}
