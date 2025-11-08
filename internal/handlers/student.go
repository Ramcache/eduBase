package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"eduBase/internal/helpers"
	"eduBase/internal/models"
	"eduBase/internal/repository"
	"eduBase/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

type StudentHandler struct {
	svc *services.StudentService
}

func NewStudentHandler(svc *services.StudentService) *StudentHandler {
	return &StudentHandler{svc: svc}
}

func (h *StudentHandler) Routes(r chi.Router) {
	r.Route("/students", func(r chi.Router) {
		r.Get("/", h.GetAll)
		r.Get("/{id}", h.GetByID)
		r.Get("/stats", h.GetStats)
		r.Get("/export", h.ExportCSV)
		r.Post("/", h.Create)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})
}

// GetByID godoc
// @Summary Получить ученика по ID
// @Tags Students
// @Produce json
// @Param id path int true "ID ученика"
// @Security BearerAuth
// @Success 200 {object} models.Student
// @Failure 403 {object} helpers.ErrorResponse
// @Failure 404 {object} helpers.ErrorResponse
// @Router /students/{id} [get]
func (h *StudentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	_, claims, _ := jwtauth.FromContext(r.Context())
	role := claims["role"].(string)
	userID := int(claims["user_id"].(float64))

	st, err := h.svc.GetByID(ctx, id)
	if err != nil {
		helpers.Error(w, http.StatusNotFound, "student not found")
		return
	}

	if role == "school" {
		schoolRepo := repository.NewSchoolRepository(h.svc.SchoolRepoDB())
		school, err := schoolRepo.GetByUserID(ctx, userID)
		if err != nil || st.SchoolID != school.ID {
			helpers.Error(w, http.StatusForbidden, "access denied")
			return
		}
	}

	helpers.JSON(w, http.StatusOK, st)
}

// Update godoc
// @Summary Обновить данные ученика
// @Tags Students
// @Accept json
// @Produce json
// @Param id path int true "ID ученика"
// @Param data body models.Student true "Новые данные"
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 400 {object} helpers.ErrorResponse
// @Failure 403 {object} helpers.ErrorResponse
// @Failure 404 {object} helpers.ErrorResponse
// @Router /students/{id} [put]
func (h *StudentHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	_, claims, _ := jwtauth.FromContext(r.Context())
	role := claims["role"].(string)
	userID := int(claims["user_id"].(float64))

	var s models.Student
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		helpers.Error(w, http.StatusBadRequest, "invalid request")
		return
	}

	if s.FullName == "" {
		helpers.Error(w, http.StatusBadRequest, "full_name required")
		return
	}

	if role == "school" {
		schoolRepo := repository.NewSchoolRepository(h.svc.SchoolRepoDB())
		school, err := schoolRepo.GetByUserID(ctx, userID)
		if err != nil {
			helpers.Error(w, http.StatusForbidden, "school not found")
			return
		}
		s.SchoolID = school.ID
	}

	ok, err := h.svc.Update(ctx, id, &s, role)
	if err != nil {
		helpers.Error(w, http.StatusInternalServerError, "failed to update student")
		return
	}
	if !ok {
		helpers.Error(w, http.StatusNotFound, "student not found or not yours")
		return
	}

	helpers.JSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// GetStats godoc
// @Summary Получить статистику по ученикам
// @Description Только для ROO (по полу, школам и т.д.)
// @Tags Students
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]int
// @Failure 403 {object} helpers.ErrorResponse
// @Router /students/stats [get]
func (h *StudentHandler) GetStats(w http.ResponseWriter, r *http.Request) {
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

// ExportCSV godoc
// @Summary Экспорт учеников в CSV (только ROO)
// @Tags Students
// @Produce text/csv
// @Security BearerAuth
// @Success 200 {string} string "csv file"
// @Router /students/export [get]
func (h *StudentHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	role := claims["role"].(string)
	if role != "roo" {
		helpers.Error(w, http.StatusForbidden, "access denied")
		return
	}

	ctx := context.Background()
	list, err := h.svc.GetAll(ctx, nil, repository.StudentFilter{})
	if err != nil {
		helpers.Error(w, http.StatusInternalServerError, "failed to export")
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=students.csv")

	fmt.Fprintln(w, "ID,Full Name,Gender,Class ID,School ID,Created At")
	for _, s := range list {
		fmt.Fprintf(w, "%d,%s,%s,%d,%d,%s\n",
			s.ID, s.FullName, deref(s.Gender), s.ClassID, s.SchoolID, s.CreatedAt.Format(time.RFC3339))
	}
}

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// GetAll godoc
// @Summary Получить список учеников
// @Tags Students
// @Produce json
// @Param full_name query string false "ФИО"
// @Param gender query string false "Пол (male/female)"
// @Param class_id query int false "ID класса"
// @Security BearerAuth
// @Success 200 {array} models.Student
// @Failure 500 {object} helpers.ErrorResponse
// @Router /students [get]
func (h *StudentHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	_, claims, _ := jwtauth.FromContext(r.Context())
	role := claims["role"].(string)
	userID := int(claims["user_id"].(float64))

	var schoolID *int
	if role == "school" {
		schoolRepo := repository.NewSchoolRepository(h.svc.SchoolRepoDB())
		school, err := schoolRepo.GetByUserID(ctx, userID)
		if err != nil {
			helpers.Error(w, http.StatusForbidden, "school not found")
			return
		}
		schoolID = &school.ID
	}

	f := repository.StudentFilter{
		FullName: r.URL.Query().Get("full_name"),
		Gender:   r.URL.Query().Get("gender"),
	}
	if v := r.URL.Query().Get("class_id"); v != "" {
		id, _ := strconv.Atoi(v)
		f.ClassID = &id
	}

	list, err := h.svc.GetAll(ctx, schoolID, f)
	if err != nil {
		helpers.Error(w, http.StatusInternalServerError, "failed to get students")
		return
	}
	helpers.JSON(w, http.StatusOK, list)
}

// Create godoc
// @Summary Добавить ученика
// @Tags Students
// @Accept json
// @Produce json
// @Param data body models.Student true "Данные ученика"
// @Security BearerAuth
// @Success 201 {object} models.Student
// @Failure 400 {object} helpers.ErrorResponse
// @Failure 403 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Router /students [post]
func (h *StudentHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	_, claims, _ := jwtauth.FromContext(r.Context())
	role := claims["role"].(string)
	userID := int(claims["user_id"].(float64))

	if role != "school" {
		helpers.Error(w, http.StatusForbidden, "only schools can add students")
		return
	}

	var s models.Student
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		helpers.Error(w, http.StatusBadRequest, "invalid request")
		return
	}

	if s.FullName == "" || s.ClassID == 0 {
		helpers.Error(w, http.StatusBadRequest, "full_name and class_id required")
		return
	}

	schoolRepo := repository.NewSchoolRepository(h.svc.SchoolRepoDB())
	school, err := schoolRepo.GetByUserID(ctx, userID)
	if err != nil {
		helpers.Error(w, http.StatusForbidden, "school not found")
		return
	}
	s.SchoolID = school.ID

	if err := h.svc.Create(ctx, &s); err != nil {
		helpers.Error(w, http.StatusInternalServerError, "failed to create student")
		return
	}
	helpers.JSON(w, http.StatusCreated, s)
}

// Delete godoc
// @Summary Удалить ученика
// @Tags Students
// @Param id path int true "ID ученика"
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 500 {object} helpers.ErrorResponse
// @Router /students/{id} [delete]
func (h *StudentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	_, claims, _ := jwtauth.FromContext(r.Context())
	userID := int(claims["user_id"].(float64))

	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	schoolRepo := repository.NewSchoolRepository(h.svc.SchoolRepoDB())
	school, err := schoolRepo.GetByUserID(ctx, userID)
	if err != nil {
		helpers.Error(w, http.StatusForbidden, "school not found")
		return
	}

	// определяем class_id для корректного обновления счётчиков
	var classID int
	h.svc.ClassRepoDB().QueryRow(ctx, `SELECT class_id FROM students WHERE id=$1`, id).Scan(&classID)

	if err := h.svc.Delete(ctx, id, school.ID, classID); err != nil {
		helpers.Error(w, http.StatusInternalServerError, "failed to delete student")
		return
	}
	helpers.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
