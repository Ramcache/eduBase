package handlers

import (
	"context"
	"eduBase/internal/repository"
	"encoding/json"
	"net/http"
	"strconv"

	"eduBase/internal/helpers"
	"eduBase/internal/models"
	"eduBase/internal/services"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

type ClassHandler struct {
	svc *services.ClassService
}

func NewClassHandler(svc *services.ClassService) *ClassHandler {
	return &ClassHandler{svc: svc}
}

func (h *ClassHandler) Routes(r chi.Router) {
	r.Route("/classes", func(r chi.Router) {
		r.Get("/", h.GetClasses)
		r.Post("/", h.Create)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})
}

// Получить список классов (ROO → все, School → свои)
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
		// 1️⃣ получаем school_id по user_id
		schoolRepo := repository.NewSchoolRepository(h.svc.RepoDB())
		school, errGet := schoolRepo.GetByUserID(ctx, userID)
		if errGet != nil {
			helpers.Error(w, http.StatusForbidden, "school not found")
			return
		}
		// 2️⃣ получаем классы этой школы
		res, err = h.svc.GetBySchool(ctx, school.ID)
	}

	if err != nil {
		helpers.Error(w, http.StatusInternalServerError, "failed to get classes")
		return
	}
	helpers.JSON(w, http.StatusOK, res)
}

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

	// Если школа — нужно получить её school_id по user_id
	if role == "school" {
		schoolRepo := repository.NewSchoolRepository(h.svc.RepoDB()) // см. ниже пояснение
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

	// Получаем school_id для роли school
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

func (h *ClassHandler) Delete(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	role := claims["role"].(string)
	userID := int(claims["user_id"].(float64))

	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	schoolID := 0
	if role == "school" {
		schoolID = userID
	}
	if err := h.svc.Delete(context.Background(), id, schoolID); err != nil {
		helpers.Error(w, http.StatusInternalServerError, "failed to delete class")
		return
	}
	helpers.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
