package handlers

import (
	"encoding/json"
	"net/http"

	"eduBase/internal/service"
)

type MedicalHandler struct{ svc service.StudentService }

func NewMedicalHandler(s service.StudentService) *MedicalHandler { return &MedicalHandler{svc: s} }

type medReq struct {
	StudentID    int     `json:"student_id"`
	Benefits     *string `json:"benefits"`
	MedicalNotes *string `json:"medical_notes"`
	HealthGroup  *int    `json:"health_group"`
	Allergies    *string `json:"allergies"`
	Activities   *string `json:"activities"`
}

// Upsert
// @Summary      Upsert medical
// @Description  Обновить/создать мед.данные/льготы/кружки
// @Tags         Medical
// @Accept       json
// @Produce      json
// @Param        payload  body      handlers.MedicalUpsertRequest  true  "Medical payload"
// @Success      200      {object}  handlers.OkResponse
// @Router       /api/students/medical [put]
func (h *MedicalHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	var req medReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", 400)
		return
	}
	err := h.svc.UpsertMedical(r.Context(), service.MedicalDTO(req))
	if err != nil {
		http.Error(w, "db error", 500)
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok"})
}
