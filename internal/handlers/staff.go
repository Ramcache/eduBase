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

	staff, err := h.svc.GetByID(ctx, id, role, userID)
	if err != nil {
		if err.Error() == "access denied" {
			helpers.Error(w, http.StatusForbidden, "access denied")
			return
		}
		helpers.Error(w, http.StatusNotFound, "staff not found")
		return
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

	ok, err := h.svc.Update(ctx, id, &s, role, userID)
	if err != nil {
		if err.Error() == "access denied" {
			helpers.Error(w, http.StatusForbidden, "access denied")
			return
		}
		helpers.Error(w, http.StatusInternalServerError, "failed to update staff")
		return
	}
	if !ok {
		helpers.Error(w, http.StatusNotFound, "staff not found")
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
	ctx := context.Background()
	_, claims, _ := jwtauth.FromContext(r.Context())
	role := claims["role"].(string)

	stats, err := h.svc.GetStats(ctx, role)
	if err != nil {
		if err.Error() == "access denied" {
			helpers.Error(w, http.StatusForbidden, "access denied")
			return
		}
		helpers.Error(w, http.StatusInternalServerError, "failed to get stats")
		return
	}

	helpers.JSON(w, http.StatusOK, stats)
}

// GetAll godoc
// @Summary Получить список сотрудников
// @Description ROO — всех, School — только своих
// @Tags Staff
// @Produce json
// @Param full_name query string false "ФИО"
// @Param phone query string false "Телефон"
// @Param position query string false "Должность"
// @Param subject query string false "Предмет"
// @Param education query string false "Образование"
// @Param category query string false "Категория"
// @Param ped_experience query int false "Минимальный пед. стаж"
// @Param total_experience query int false "Минимальный общий стаж"
// @Param limit query int false "Лимит на страницу"
// @Param offset query int false "Смещение"
// @Security BearerAuth
// @Success 200 {array} models.Staff
// @Failure 500 {object} helpers.ErrorResponse
// @Router /staff [get]
func (h *StaffHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	_, claims, _ := jwtauth.FromContext(r.Context())
	role := claims["role"].(string)
	userID := int(claims["user_id"].(float64))

	q := r.URL.Query()

	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	var pedExpPtr, totalExpPtr *int
	if v := q.Get("ped_experience"); v != "" {
		if x, err := strconv.Atoi(v); err == nil {
			pedExpPtr = &x
		}
	}
	if v := q.Get("total_experience"); v != "" {
		if x, err := strconv.Atoi(v); err == nil {
			totalExpPtr = &x
		}
	}

	filter := repository.StaffFilter{
		FullName:        q.Get("full_name"),
		Phone:           q.Get("phone"),
		Position:        q.Get("position"),
		Subject:         q.Get("subject"),
		Education:       q.Get("education"),
		Category:        q.Get("category"),
		PedExperience:   pedExpPtr,
		TotalExperience: totalExpPtr,
		Limit:           limit,
		Offset:          offset,
	}

	list, err := h.svc.GetAll(ctx, role, userID, filter)
	if err != nil {
		if err.Error() == "access denied" {
			helpers.Error(w, http.StatusForbidden, "access denied")
			return
		}
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

	if err := h.svc.Create(ctx, &s, role, userID); err != nil {
		if err.Error() == "access denied" {
			helpers.Error(w, http.StatusForbidden, "access denied")
			return
		}
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
// @Failure 404 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Router /staff/{id} [delete]
func (h *StaffHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	_, claims, _ := jwtauth.FromContext(r.Context())
	role := claims["role"].(string)
	userID := int(claims["user_id"].(float64))

	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	ok, err := h.svc.Delete(ctx, id, role, userID)
	if err != nil {
		if err.Error() == "access denied" {
			helpers.Error(w, http.StatusForbidden, "access denied")
			return
		}
		helpers.Error(w, http.StatusInternalServerError, "failed to delete staff")
		return
	}
	if !ok {
		helpers.Error(w, http.StatusNotFound, "staff not found")
		return
	}

	helpers.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
