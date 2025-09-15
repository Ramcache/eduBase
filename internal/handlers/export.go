package handlers

import (
	"net/http"
	"strconv"
	"time"

	"eduBase/internal/repository"
	"eduBase/internal/service"
	"eduBase/internal/utils/excel"
)

type ExportHandler struct{ svc service.StudentService }

func NewExportHandler(s service.StudentService) *ExportHandler { return &ExportHandler{svc: s} }

// ExportStudents
/*
@Summary      Export students (full) to Excel
@Description  Полный экспорт: Core + Documents + Medical + Consents + Contacts
@Tags         Export
@Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
@Param        q                     query    string  false  "ILIKE ФИО"
@Param        school_id             query    int     false  "ID школы"
@Param        class                 query    string  false  "Класс (например 7А)"
@Param        status                query    string  false  "enrolled|transferred|graduated|expelled"
@Param        admission_year_from   query    int     false  "Год поступления с"
@Param        admission_year_to     query    int     false  "Год поступления по"
@Param        birth_date_from       query    string  false  "Дата рождения c (DD-MM-YYYY)"
@Param        birth_date_to         query    string  false  "Дата рождения по (DD-MM-YYYY)"
@Success      200  {file}  file  "Excel file"
@Router       /api/students/export.xlsx [get]
*/
func (h *ExportHandler) ExportStudents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	classLabel := r.URL.Query().Get("class")
	status := r.URL.Query().Get("status")
	schoolID := parseIntPtr(r.URL.Query().Get("school_id"))

	ayFrom := parseIntPtr(r.URL.Query().Get("admission_year_from"))
	ayTo := parseIntPtr(r.URL.Query().Get("admission_year_to"))
	bdFrom := parseDatePtr(r.URL.Query().Get("birth_date_from"))
	bdTo := parseDatePtr(r.URL.Query().Get("birth_date_to"))

	var classPtr, statusPtr *string
	if classLabel != "" {
		classPtr = &classLabel
	}
	if status != "" {
		statusPtr = &status
	}

	f := repository.StudentFilters{
		Q:                 q,
		SchoolID:          schoolID,
		ClassLabel:        classPtr,
		Status:            statusPtr,
		AdmissionYearFrom: ayFrom,
		AdmissionYearTo:   ayTo,
		BirthDateFrom:     bdFrom,
		BirthDateTo:       bdTo,
	}

	students, dmap, mmap, cmap, emap, _, err := h.svc.CollectExportData(r.Context(), f, 50000)
	if err != nil {
		http.Error(w, "db error", 500)
		return
	}

	xl, err := excel.BuildStudentsWorkbookFull(students, dmap, mmap, cmap, emap)
	if err != nil {
		http.Error(w, "excel error", 500)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", `attachment; filename="students_full_export.xlsx"`)
	_ = xl.Write(w)
}

func parseIntPtr(s string) *int {
	if s == "" {
		return nil
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &v
}
func parseDatePtr(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse("02.01.2006", s)
	if err != nil {
		return nil
	}
	return &t
}
