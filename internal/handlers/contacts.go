package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"eduBase/internal/service"
	"github.com/go-chi/chi/v5"
)

type ContactsHandler struct{ svc service.StudentService }

func NewContactsHandler(s service.StudentService) *ContactsHandler { return &ContactsHandler{svc: s} }

type contactReq struct {
	StudentID int    `json:"student_id"`
	FullName  string `json:"full_name"`
	Phone     string `json:"phone"`
	Relation  string `json:"relation"`
}

// Add
// @Summary      Add emergency contact
// @Description  Добавить экстренный контакт к ученику
// @Tags         Contacts
// @Accept       json
// @Produce      json
// @Param        id       path      int                       true  "Student ID (в URL)"
// @Param        payload  body      handlers.ContactAddRequest true  "Contact payload (содержит student_id)"
// @Success      201      {object}  handlers.CreateIDResponse
// @Router       /api/students/{id}/contacts [post]
func (h *ContactsHandler) Add(w http.ResponseWriter, r *http.Request) {
	var req contactReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", 400)
		return
	}
	id, err := h.svc.AddContact(r.Context(), service.ContactDTO(req))
	if err != nil {
		http.Error(w, "db error", 500)
		return
	}
	writeJSON(w, 201, map[string]any{"id": id})
}

// Delete
// @Summary      Delete emergency contact
// @Description  Удалить экстренный контакт по ID
// @Tags         Contacts
// @Produce      json
// @Param        id   path  int  true  "Contact ID"
// @Success      200  {object}  handlers.OkResponse
// @Router       /api/students/contacts/{id} [delete]
func (h *ContactsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := h.svc.DeleteContact(r.Context(), id); err != nil {
		http.Error(w, "db error", 500)
		return
	}
	writeJSON(w, 200, map[string]string{"status": "deleted"})
}
