package handlers

import (
	"encoding/json"
	"net/http"

	"eduBase/internal/service"
)

type ConsentsHandler struct{ svc service.StudentService }

func NewConsentsHandler(s service.StudentService) *ConsentsHandler { return &ConsentsHandler{svc: s} }

type consReq struct {
	StudentID           int     `json:"student_id"`
	ConsentPD           bool    `json:"consent_data_processing"`
	ConsentPDDate       *string `json:"consent_data_processing_date"`
	ConsentPhoto        bool    `json:"consent_photo_publication"`
	ConsentPhotoDate    *string `json:"consent_photo_publication_date"`
	ConsentInternet     bool    `json:"consent_internet_access"`
	ConsentInternetDate *string `json:"consent_internet_access_date"`
}

// Upsert
// @Summary      Upsert consents
// @Description  Обновить/создать согласия (ПДн, фото, интернет)
// @Tags         Consents
// @Accept       json
// @Produce      json
// @Param        payload  body      handlers.ConsentsUpsertRequest  true  "Consents payload"
// @Success      200      {object}  handlers.OkResponse
// @Failure      422      {string}  string  "validation error"
// @Router       /api/students/consents [put]
func (h *ConsentsHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	var req consReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", 400)
		return
	}
	err := h.svc.UpsertConsents(r.Context(), service.ConsentsDTO(req))
	if err != nil {
		http.Error(w, err.Error(), 422)
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok"})
}
