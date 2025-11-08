package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"eduBase/internal/helpers"
	"eduBase/internal/models"
	"eduBase/internal/services"
	"github.com/go-chi/chi/v5"
)

type RooSchoolHandler struct {
	svc *services.SchoolService
}

func NewRooSchoolHandler(svc *services.SchoolService) *RooSchoolHandler {
	return &RooSchoolHandler{svc: svc}
}

func (h *RooSchoolHandler) Routes(r chi.Router) {
	r.Route("/roo/schools", func(r chi.Router) {
		r.Get("/", h.GetAll)
		r.Get("/{id}", h.GetByID)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})
}

func (h *RooSchoolHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.GetAll(context.Background())
	if err != nil {
		helpers.Error(w, http.StatusInternalServerError, "failed to load schools")
		return
	}
	helpers.JSON(w, http.StatusOK, list)
}

func (h *RooSchoolHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	school, err := h.svc.GetByID(context.Background(), id)
	if err != nil {
		helpers.Error(w, http.StatusNotFound, "school not found")
		return
	}
	helpers.JSON(w, http.StatusOK, school)
}

func (h *RooSchoolHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var req models.School
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.Update(context.Background(), id, &req); err != nil {
		helpers.Error(w, http.StatusInternalServerError, "failed to update school")
		return
	}
	helpers.JSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *RooSchoolHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := h.svc.Delete(context.Background(), id); err != nil {
		helpers.Error(w, http.StatusInternalServerError, "failed to delete school")
		return
	}
	helpers.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
