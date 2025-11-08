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

// GetAll godoc
// @Summary      Получить все школы
// @Description  Возвращает список всех школ (только для ROO)
// @Tags         Schools
// @Produce      json
// @Success      200 {array} models.School
// @Failure      500 {object} helpers.ErrorResponse
// @Security     BearerAuth
// @Router       /roo/schools [get]
func (h *RooSchoolHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.GetAll(context.Background())
	if err != nil {
		helpers.Error(w, http.StatusInternalServerError, "failed to load schools")
		return
	}
	helpers.JSON(w, http.StatusOK, list)
}

// GetByID godoc
// @Summary      Получить школу по ID
// @Description  Возвращает данные конкретной школы (только для ROO)
// @Tags         Schools
// @Produce      json
// @Param        id path int true "ID школы"
// @Success      200 {object} models.School
// @Failure      404 {object} helpers.ErrorResponse
// @Failure      500 {object} helpers.ErrorResponse
// @Security     BearerAuth
// @Router       /roo/schools/{id} [get]
func (h *RooSchoolHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	school, err := h.svc.GetByID(context.Background(), id)
	if err != nil {
		helpers.Error(w, http.StatusNotFound, "school not found")
		return
	}
	helpers.JSON(w, http.StatusOK, school)
}

// Update godoc
// @Summary      Обновить школу
// @Description  Обновляет данные школы по ID (только для ROO)
// @Tags         Schools
// @Accept       json
// @Produce      json
// @Param        id path int true "ID школы"
// @Param        request body models.School true "Поля для обновления"
// @Success      200 {object} map[string]string
// @Failure      400 {object} helpers.ErrorResponse
// @Failure      500 {object} helpers.ErrorResponse
// @Security     BearerAuth
// @Router       /roo/schools/{id} [put]
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

// Delete godoc
// @Summary      Удалить школу
// @Description  Удаляет школу по ID (только для ROO)
// @Tags         Schools
// @Produce      json
// @Param        id path int true "ID школы"
// @Success      200 {object} map[string]string
// @Failure      500 {object} helpers.ErrorResponse
// @Security     BearerAuth
// @Router       /roo/schools/{id} [delete]
func (h *RooSchoolHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := h.svc.Delete(context.Background(), id); err != nil {
		helpers.Error(w, http.StatusInternalServerError, "failed to delete school")
		return
	}
	helpers.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
