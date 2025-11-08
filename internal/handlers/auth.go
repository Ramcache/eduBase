package handlers

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"net/http"

	"eduBase/internal/helpers"
	"eduBase/internal/services"
)

type AuthHandler struct {
	svc *services.AuthService
}

func NewAuthHandler(svc *services.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Routes(r chi.Router) {
	r.Post("/auth/login", h.Login)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	token, err := h.svc.Login(context.Background(), req.Email, req.Password)
	if err != nil {
		helpers.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	helpers.JSON(w, http.StatusOK, map[string]string{"token": token})
}
