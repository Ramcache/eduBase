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

// ClassHandler — обработчик классов
type ClassHandler struct {
	svc *services.ClassService
}

func NewClassHandler(svc *services.ClassService) *ClassHandler {
	return &ClassHandler{svc: svc}
}

func (h *ClassHandler) Routes(r chi.Router) {
	r.Route("/classes", func(r chi.Router) {
		r.Get("/", h.GetClasses)
		r.Get("/{id}", h.GetByID)
		r.Post("/", h.Create)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})
}

// GetClasses godoc
// @Summary Получить список классов
// @Description ROO — все классы, School — только свои
// @Tags Classes
// @Produce json
// @Success 200 {array} models.Class
// @Failure 403 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Security BearerAuth
// @Router /classes [get]
func (h *ClassHandler) GetClasses(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	_, claims, _ := jwtauth.FromContext(r.Context())
	role := claims["role"].(string)
	userID := int(claims["user_id"].(float64))

	var res []models.Class
	var err error

	if role == "roo" {
		res, err = h.svc.GetAll(ctx)
	} else if role == "school" {
		schoolRepo := repository.NewSchoolRepository(h.svc.RepoDB())
		school, errGet := schoolRepo.GetByUserID(ctx, userID)
		if errGet != nil {
			helpers.Error(w, http.StatusForbidden, "school not found")
			return
		}
		res, err = h.svc.GetBySchool(ctx, school.ID)
	}

	if err != nil {
		helpers.Error(w, http.StatusInternalServerError, "failed to get classes")
		return
	}
	helpers.JSON(w, http.StatusOK, res)
}

// Create godoc
// @Summary Создать новый класс
// @Description School создаёт свой класс
// @Tags Classes
// @Accept json
// @Produce json
// @Param data body models.Class true "Данные класса"
// @Success 201 {object} models.Class
// @Failure 400 {object} helpers.ErrorResponse
// @Failure 403 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Security BearerAuth
// @Router /classes [post]
func (h *ClassHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	_, claims, _ := jwtauth.FromContext(r.Context())
	role := claims["role"].(string)
	userID := int(claims["user_id"].(float64))

	var c models.Class
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		helpers.Error(w, http.StatusBadRequest, "invalid request")
		return
	}

	if role == "school" {
		schoolRepo := repository.NewSchoolRepository(h.svc.RepoDB())
		school, err := schoolRepo.GetByUserID(ctx, userID)
		if err != nil {
			helpers.Error(w, http.StatusForbidden, "school not found")
			return
		}
		c.SchoolID = school.ID
	}

	if err := h.svc.Create(ctx, &c); err != nil {
		helpers.Error(w, http.StatusInternalServerError, "failed to create class")
		return
	}
	helpers.JSON(w, http.StatusCreated, c)
}

// Update godoc
// @Summary Обновить класс
// @Description ROO — может обновить любой, School — только свой
// @Tags Classes
// @Accept json
// @Produce json
// @Param id path int true "ID класса"
// @Param data body models.Class true "Данные для обновления"
// @Success 200 {object} map[string]string
// @Failure 400 {object} helpers.ErrorResponse
// @Failure 403 {object} helpers.ErrorResponse
// @Failure 404 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Security BearerAuth
// @Router /classes/{id} [put]
func (h *ClassHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	_, claims, _ := jwtauth.FromContext(r.Context())
	role := claims["role"].(string)
	userID := int(claims["user_id"].(float64))

	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var c models.Class
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		helpers.Error(w, http.StatusBadRequest, "invalid request")
		return
	}

	if role == "school" {
		schoolRepo := repository.NewSchoolRepository(h.svc.RepoDB())
		school, err := schoolRepo.GetByUserID(ctx, userID)
		if err != nil {
			helpers.Error(w, http.StatusForbidden, "school not found")
			return
		}
		c.SchoolID = school.ID
	}

	ok, err := h.svc.Update(ctx, id, &c, role)
	if err != nil {
		helpers.Error(w, http.StatusInternalServerError, "failed to update class")
		return
	}
	if !ok {
		helpers.Error(w, http.StatusNotFound, "class not found or not yours")
		return
	}

	helpers.JSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// Delete godoc
// @Summary Удалить класс
// @Description School — только свои, ROO — любые
// @Tags Classes
// @Param id path int true "ID класса"
// @Success 200 {object} map[string]string
// @Failure 403 {object} helpers.ErrorResponse
// @Failure 404 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Security BearerAuth
// @Router /classes/{id} [delete]
func (h *ClassHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
		helpers.Error(w, http.StatusInternalServerError, "failed to delete class")
		return
	}
	helpers.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// GetByID godoc
// @Summary Получить класс по ID
// @Description ROO — любой класс, School — только свой
// @Tags Classes
// @Produce json
// @Param id path int true "ID класса"
// @Success 200 {object} models.Class
// @Failure 403 {object} helpers.ErrorResponse
// @Failure 404 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Security BearerAuth
// @Router /classes/{id} [get]
func (h *ClassHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	_, claims, _ := jwtauth.FromContext(r.Context())
	role := claims["role"].(string)
	userID := int(claims["user_id"].(float64))

	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	class, err := h.svc.GetByID(ctx, id)
	if err != nil {
		helpers.Error(w, http.StatusNotFound, "class not found")
		return
	}

	// Проверка доступа: если School, то только свой класс
	if role == "school" {
		schoolRepo := repository.NewSchoolRepository(h.svc.RepoDB())
		school, err := schoolRepo.GetByUserID(ctx, userID)
		if err != nil {
			helpers.Error(w, http.StatusForbidden, "school not found")
			return
		}
		if class.SchoolID != school.ID {
			helpers.Error(w, http.StatusForbidden, "access denied")
			return
		}
	}

	helpers.JSON(w, http.StatusOK, class)
}
