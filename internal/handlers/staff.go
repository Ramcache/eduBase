package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"eduBase/internal/helpers"
	"eduBase/internal/models"
	"eduBase/internal/repository"
	"eduBase/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

type StaffHandler struct {
	svc *services.StaffService
}

func NewStaffHandler(svc *services.StaffService) *StaffHandler {
	return &StaffHandler{svc: svc}
}

func (h *StaffHandler) Routes(r chi.Router) {
	r.Route("/staff", func(r chi.Router) {
		r.Get("/", h.GetAll)
		r.Get("/{id}", h.GetByID)
		r.Get("/stats", h.GetStats)
		r.Post("/", h.Create)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})
}

// GetByID godoc
// @Summary Получить сотрудника по ID
// @Description ROO — любого, School — только своего
// @Tags Staff
// @Produce json
// @Param id path int true "ID сотрудника"
// @Security BearerAuth
// @Success 200 {object} models.Staff
// @Failure 403 {object} helpers.ErrorResponse
// @Failure 404 {object} helpers.ErrorResponse
// @Router /staff/{id} [get]
func (h *StaffHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	_, claims, _ := jwtauth.FromContext(r.Context())
	role := claims["role"].(string)
	userID := int(claims["user_id"].(float64))

	staff, err := h.svc.GetByID(ctx, id)
	if err != nil {
		helpers.Error(w, http.StatusNotFound, "staff not found")
		return
	}

	if role == "school" {
		schoolRepo := repository.NewSchoolRepository(h.svc.RepoDB())
		school, err := schoolRepo.GetByUserID(ctx, userID)
		if err != nil || staff.SchoolID != school.ID {
			helpers.Error(w, http.StatusForbidden, "access denied")
			return
		}
	}

	helpers.JSON(w, http.StatusOK, staff)
}

// Update godoc
// @Summary Обновить данные сотрудника
// @Description ROO — любого, School — только своего
// @Tags Staff
// @Accept json
// @Produce json
// @Param id path int true "ID сотрудника"
// @Param data body models.Staff true "Обновлённые данные"
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 403 {object} helpers.ErrorResponse
// @Failure 404 {object} helpers.ErrorResponse
// @Router /staff/{id} [put]
func (h *StaffHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	_, claims, _ := jwtauth.FromContext(r.Context())
	role := claims["role"].(string)
	userID := int(claims["user_id"].(float64))

	var s models.Staff
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		helpers.Error(w, http.StatusBadRequest, "invalid request")
		return
	}

	if s.FullName == "" || s.Phone == "" || s.Position == "" {
		helpers.Error(w, http.StatusBadRequest, "full_name, phone and position required")
		return
	}

	if role == "school" {
		schoolRepo := repository.NewSchoolRepository(h.svc.RepoDB())
		school, err := schoolRepo.GetByUserID(ctx, userID)
		if err != nil {
			helpers.Error(w, http.StatusForbidden, "school not found")
			return
		}
		s.SchoolID = school.ID
	}

	ok, err := h.svc.Update(ctx, id, &s, role)
	if err != nil {
		helpers.Error(w, http.StatusInternalServerError, "failed to update staff")
		return
	}
	if !ok {
		helpers.Error(w, http.StatusNotFound, "staff not found or not yours")
		return
	}

	helpers.JSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// GetStats godoc
// @Summary Получить статистику по персоналу (ROO)
// @Description Кол-во сотрудников по должностям
// @Tags Staff
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]int
// @Failure 403 {object} helpers.ErrorResponse
// @Router /staff/stats [get]
func (h *StaffHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	role := claims["role"].(string)
	if role != "roo" {
		helpers.Error(w, http.StatusForbidden, "access denied")
		return
	}
	stats, err := h.svc.GetStats(context.Background())
	if err != nil {
		helpers.Error(w, http.StatusInternalServerError, "failed to get stats")
		return
	}
	helpers.JSON(w, http.StatusOK, stats)
}

func (h *StaffHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	_, claims, _ := jwtauth.FromContext(r.Context())
	role := claims["role"].(string)
	userID := int(claims["user_id"].(float64))

	var schoolID *int

	// 🔒 Школа видит только свои данные
	if role == "school" {
		schoolRepo := repository.NewSchoolRepository(h.svc.RepoDB())
		school, err := schoolRepo.GetByUserID(ctx, userID)
		if err != nil {
			helpers.Error(w, http.StatusForbidden, "school not found")
			return
		}
		schoolID = &school.ID
	}

	if role != "roo" && role != "school" {
		helpers.Error(w, http.StatusForbidden, "access denied")
		return
	}

	// Фильтры
	filter := repository.StaffFilter{
		FullName:  r.URL.Query().Get("full_name"),
		Phone:     r.URL.Query().Get("phone"),
		Position:  r.URL.Query().Get("position"),
		Subject:   r.URL.Query().Get("subject"),
		Education: r.URL.Query().Get("education"),
		Category:  r.URL.Query().Get("category"),
	}

	list, err := h.svc.GetAll(ctx, schoolID, filter)
	if err != nil {
		helpers.Error(w, http.StatusInternalServerError, "failed to get staff")
		return
	}

	helpers.JSON(w, http.StatusOK, list)
}

// Create godoc
// @Summary Добавить сотрудника
// @Description Только школа может добавлять своих сотрудников
// @Tags Staff
// @Accept json
// @Produce json
// @Param data body models.Staff true "Данные сотрудника"
// @Security BearerAuth
// @Success 201 {object} models.Staff
// @Failure 400 {object} helpers.ErrorResponse
// @Failure 403 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Router /staff [post]
func (h *StaffHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	_, claims, _ := jwtauth.FromContext(r.Context())
	role := claims["role"].(string)
	userID := int(claims["user_id"].(float64))

	if role != "school" {
		helpers.Error(w, http.StatusForbidden, "only schools can add staff")
		return
	}

	var s models.Staff
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		helpers.Error(w, http.StatusBadRequest, "invalid request")
		return
	}

	if s.FullName == "" || s.Phone == "" || s.Position == "" {
		helpers.Error(w, http.StatusBadRequest, "full_name, phone and position are required")
		return
	}

	schoolRepo := repository.NewSchoolRepository(h.svc.RepoDB())
	school, err := schoolRepo.GetByUserID(ctx, userID)
	if err != nil {
		helpers.Error(w, http.StatusForbidden, "school not found")
		return
	}
	s.SchoolID = school.ID

	if err := h.svc.Create(ctx, &s); err != nil {
		helpers.Error(w, http.StatusInternalServerError, "failed to create staff")
		return
	}

	helpers.JSON(w, http.StatusCreated, s)
}

// Delete godoc
// @Summary Удалить сотрудника
// @Description School — только своего, ROO — любого
// @Tags Staff
// @Param id path int true "ID сотрудника"
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 403 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Router /staff/{id} [delete]
func (h *StaffHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	_, claims, _ := jwtauth.FromContext(r.Context())
	role := claims["role"].(string)
	userID := int(claims["user_id"].(float64))

	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	schoolID := 0

	if role == "school" {
		schoolRepo := repository.NewSchoolRepository(h.svc.RepoDB())
		school, err := schoolRepo.GetByUserID(ctx, userID)
		if err != nil {
			helpers.Error(w, http.StatusForbidden, "school not found")
			return
		}
		schoolID = school.ID
	}

	if err := h.svc.Delete(ctx, id, schoolID); err != nil {
		helpers.Error(w, http.StatusInternalServerError, "failed to delete staff")
		return
	}
	helpers.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
