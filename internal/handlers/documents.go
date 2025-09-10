package handlers

import (
	"encoding/json"
	"net/http"

	"eduBase/internal/service"
)

type DocumentsHandler struct{ svc service.StudentService }

func NewDocumentsHandler(s service.StudentService) *DocumentsHandler {
	return &DocumentsHandler{svc: s}
}

type docsReq struct {
	StudentID        int     `json:"student_id"`
	SNILS            string  `json:"snils"`
	PassportSeries   *string `json:"passport_series"`
	PassportNumber   *string `json:"passport_number"`
	BirthCertificate *string `json:"birth_certificate"`
	BirthDate        string  `json:"birth_date"`
}

// Upsert
// @Summary      Upsert documents
// @Description  Обновить/создать документы ученика (СНИЛС обязателен; паспорт обязателен с 14 лет и 1 месяц, иначе — свидетельство)
// @Tags         Documents
// @Accept       json
// @Produce      json
// @Param        payload  body      handlers.DocumentsUpsertRequest  true  "Documents payload"
// @Success      200      {object}  handlers.OkResponse
// @Failure      422      {string}  string  "validation error"
// @Router       /api/students/docs [put]
func (h *DocumentsHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	var req docsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", 400)
		return
	}
	err := h.svc.UpsertDocuments(r.Context(), service.DocumentsDTO(req))
	if err != nil {
		http.Error(w, err.Error(), 422)
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok"})
}
