package handlers

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"time"

	tpl "eduBase/internal/assets/templates"
	"eduBase/internal/service"
	excelutil "eduBase/internal/utils/excel"
	"github.com/xuri/excelize/v2"
)

type ImportHandler struct{ svc service.StudentService }

func NewImportHandler(s service.StudentService) *ImportHandler { return &ImportHandler{svc: s} }

// Template
/*
@Summary      Download import template
@Description  Шаблон Excel: листы Students/Contacts
@Tags         Import
@Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
@Success      200  {file}  file  "Excel template"
@Router       /api/students/import/template.xlsx [get]
*/
func (h *ImportHandler) Template(w http.ResponseWriter, r *http.Request) {
	b, err := tpl.Bytes()
	if err != nil {
		http.Error(w, "template not found", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", `attachment; filename="students_import_template.xlsx"`)
	w.Header().Set("Cache-Control", "public, max-age=86400")

	// красиво отдадим как файл (поддержка Range/кешей у клиентов)
	http.ServeContent(w, r, "students_import_template.xlsx", time.Time{}, bytes.NewReader(b))
}

// Import
/*
@Summary      Import students from Excel
@Description  Импорт из ОДНОГО листа `Students`. В конце листа есть колонки Контакт1..3 (ФИО/Телефон/Связь). Для совместимости можно также добавить второй лист `Contacts` — он будет прочитан дополнительно.
@Tags         Import
@Accept       mpfd
@Produce      json
@Param        file  formData  file   true   "Excel .xlsx file"
@Param        replace_contacts query bool false "Удалять старые контакты и заменять из файла (default: false)"
@Success      200  {array}   service.ImportResult
@Failure      400  {string}  string  "bad request"
@Router       /api/students/import.xlsx [post]
*/
func (h *ImportHandler) Import(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(64 << 20); err != nil {
		http.Error(w, "bad form", 400)
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file required", 400)
		return
	}
	defer file.Close()

	// читаем в excelize
	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "read error", 400)
		return
	}
	xl, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		http.Error(w, "xlsx open error", 400)
		return
	}

	rows, contacts, err := excelutil.ParseImportWorkbook(xl)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	replace := false
	if v := r.URL.Query().Get("replace_contacts"); v != "" {
		b, _ := strconv.ParseBool(v)
		replace = b
	}

	results, err := h.svc.Import(r.Context(), rows, contacts, userID(r), replace)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	writeJSON(w, 200, results)
}
