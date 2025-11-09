package handlers

import (
	"context"
	"net/http"
	"strconv"

	"eduBase/internal/helpers"
	"eduBase/internal/repository"
	"eduBase/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

// StatsHandler — агрегаты по классам/ученикам/учителям
type StatsHandler struct {
	svc *services.StatsService
}

func NewStatsHandler(svc *services.StatsService) *StatsHandler {
	return &StatsHandler{svc: svc}
}

func (h *StatsHandler) Routes(r chi.Router) {
	r.Route("/stats", func(r chi.Router) {
		r.Get("/summary", h.Summary)
	})
}

// Summary godoc
// @Summary Сводная статистика (кол-во классов, учеников, учителей)
// @Description ROO — вся система или по school_id; School — только своя школа (параметр игнорируется)
// @Tags Stats
// @Produce json
// @Param school_id query int false "Фильтрация по школе (только для ROO)"
// @Security BearerAuth
// @Success 200 {object} models.StatsSummary "schools, classes, students, teachers, staff_total"
// @Failure 400 {object} helpers.ErrorResponse
// @Failure 403 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Router /stats/summary [get]
func (h *StatsHandler) Summary(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	_, claims, _ := jwtauth.FromContext(r.Context())
	role := claims["role"].(string)
	userID := int(claims["user_id"].(float64))

	var schoolID *int

	// ROO: может запросить ?school_id=..., иначе вся система
	if role == "roo" {
		if v := r.URL.Query().Get("school_id"); v != "" {
			id, err := strconv.Atoi(v)
			if err != nil || id <= 0 {
				helpers.Error(w, http.StatusBadRequest, "invalid school_id")
				return
			}
			// валидация наличия школы
			exists, err := h.svc.RepoDB().Query(ctx, `SELECT 1`) // ping-like
			if err != nil {
				helpers.Error(w, http.StatusInternalServerError, "db error")
				return
			}
			exists.Close() // просто закрываем курсор, без присваивания

			ok, err := repository.NewStatsRepository(h.svc.RepoDB()).SchoolExists(ctx, id)
			if err != nil {
				helpers.Error(w, http.StatusInternalServerError, "db error")
				return
			}
			if !ok {
				helpers.Error(w, http.StatusBadRequest, "school not found")
				return
			}
			schoolID = &id
		}
	} else if role == "school" {
		// School: игнорируем переданный school_id, берём свой по user_id
		sRepo := repository.NewSchoolRepository(h.svc.RepoDB())
		school, err := sRepo.GetByUserID(ctx, userID)
		if err != nil {
			helpers.Error(w, http.StatusForbidden, "school not found")
			return
		}
		schoolID = &school.ID
	} else {
		helpers.Error(w, http.StatusForbidden, "access denied")
		return
	}

	res, err := h.svc.GetSummary(ctx, schoolID)
	if err != nil {
		helpers.Error(w, http.StatusInternalServerError, "failed to get stats")
		return
	}
	helpers.JSON(w, http.StatusOK, res)
}
