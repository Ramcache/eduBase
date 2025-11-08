package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"eduBase/internal/helpers"
	"eduBase/internal/services"
	"github.com/go-chi/chi/v5"
)

type RooHandler struct {
	svc *services.AuthService
}

func NewRooHandler(svc *services.AuthService) *RooHandler {
	return &RooHandler{svc: svc}
}

func (h *RooHandler) Routes(r chi.Router) {
	r.Route("/roo", func(r chi.Router) {
		r.Post("/register_school", h.RegisterSchool)
	})
}

type registerSchoolRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *RooHandler) RegisterSchool(w http.ResponseWriter, r *http.Request) {
	var req registerSchoolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.svc.RegisterSchool(context.Background(), req.Email, req.Password); err != nil {
		helpers.Error(w, http.StatusInternalServerError, "failed to register school")
		return
	}

	helpers.JSON(w, http.StatusCreated, map[string]string{"status": "school registered"})
}
