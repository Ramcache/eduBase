package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/xuri/excelize/v2"
	"net/http"
	"strconv"
	"time"

	"eduBase/internal/helpers"
	"eduBase/internal/models"
	"eduBase/internal/repository"
	"eduBase/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

type StudentHandler struct {
	svc *services.StudentService
}

func NewStudentHandler(svc *services.StudentService) *StudentHandler {
	return &StudentHandler{svc: svc}
}

func (h *StudentHandler) Routes(r chi.Router) {
	r.Route("/students", func(r chi.Router) {
		r.Get("/", h.GetAll)
		r.Get("/{id}", h.GetByID)
		r.Get("/stats", h.GetStats)
		r.Get("/grade-stats", h.GetGradeStats)
		r.Get("/age-stats", h.GetAgeStats)
		r.Get("/import/template", h.ImportTemplate)
		r.Get("/export", h.ExportCSV)
		r.Post("/", h.Create)
		r.Post("/import", h.ImportExcel)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})
}

// GetByID godoc
// @Summary Получить ученика по ID
// @Tags Students
// @Produce json
// @Param id path int true "ID ученика"
// @Security BearerAuth
// @Success 200 {object} models.Student
// @Failure 403 {object} helpers.ErrorResponse
// @Failure 404 {object} helpers.ErrorResponse
// @Router /students/{id} [get]
func (h *StudentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	_, claims, _ := jwtauth.FromContext(r.Context())
	role := claims["role"].(string)
	userID := int(claims["user_id"].(float64))

	st, err := h.svc.GetByID(ctx, id, role, userID)
	if err != nil {
		if err.Error() == "access denied" {
			helpers.Error(w, http.StatusForbidden, "access denied")
			return
		}
		helpers.Error(w, http.StatusNotFound, "student not found")
		return
	}

	helpers.JSON(w, http.StatusOK, st)
}

// Update godoc
// @Summary Обновить данные ученика
// @Tags Students
// @Accept json
// @Produce json
// @Param id path int true "ID ученика"
// @Param data body models.Student true "Новые данные"
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 400 {object} helpers.ErrorResponse
// @Failure 403 {object} helpers.ErrorResponse
// @Failure 404 {object} helpers.ErrorResponse
// @Router /students/{id} [put]
func (h *StudentHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	_, claims, _ := jwtauth.FromContext(r.Context())
	role := claims["role"].(string)
	userID := int(claims["user_id"].(float64))

	var s models.Student
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		helpers.Error(w, http.StatusBadRequest, "invalid request")
		return
	}

	if s.FullName == "" {
		helpers.Error(w, http.StatusBadRequest, "full_name required")
		return
	}

	ok, err := h.svc.Update(ctx, id, &s, role, userID)
	if err != nil {
		if err.Error() == "access denied" {
			helpers.Error(w, http.StatusForbidden, "access denied")
			return
		}
		helpers.Error(w, http.StatusInternalServerError, "failed to update student")
		return
	}
	if !ok {
		helpers.Error(w, http.StatusNotFound, "student not found")
		return
	}

	helpers.JSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// GetStats godoc
// @Summary Получить статистику по ученикам
// @Description Только для ROO (по полу, школам и т.д.)
// @Tags Students
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]int
// @Failure 403 {object} helpers.ErrorResponse
// @Router /students/stats [get]
func (h *StudentHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	_, claims, _ := jwtauth.FromContext(r.Context())
	role := claims["role"].(string)

	stats, err := h.svc.GetStats(ctx, role)
	if err != nil {
		if err.Error() == "access denied" {
			helpers.Error(w, http.StatusForbidden, "access denied")
			return
		}
		helpers.Error(w, http.StatusInternalServerError, "failed to get stats")
		return
	}
	helpers.JSON(w, http.StatusOK, stats)
}

// ExportCSV godoc
// @Summary Экспорт учеников в CSV (только ROO)
// @Tags Students
// @Produce text/csv
// @Security BearerAuth
// @Success 200 {string} string "csv file"
// @Router /students/export [get]
func (h *StudentHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	_, claims, _ := jwtauth.FromContext(r.Context())
	role := claims["role"].(string)
	userID := int(claims["user_id"].(float64))
	if role != "roo" {
		helpers.Error(w, http.StatusForbidden, "access denied")
		return
	}

	list, err := h.svc.GetAll(ctx, role, userID, repository.StudentFilter{})
	if err != nil {
		helpers.Error(w, http.StatusInternalServerError, "failed to export")
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=students.csv")

	fmt.Fprintln(w, "ID,Full Name,Gender,Class ID,School ID,Created At")
	for _, s := range list {
		fmt.Fprintf(w, "%d,%s,%s,%d,%d,%s\n",
			s.ID, s.FullName, deref(s.Gender), s.ClassID, s.SchoolID, s.CreatedAt.Format(time.RFC3339))
	}
}

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// GetAll godoc
// @Summary Получить список учеников
// @Tags Students
// @Produce json
// @Param full_name   query string false "ФИО"
// @Param gender      query string false "Пол (male/female)"
// @Param class_id    query int    false "ID класса"
// @Param grade_from  query int    false "Нижняя граница класса (номер)"
// @Param grade_to    query int    false "Верхняя граница класса (номер)"
// @Param age_from    query int    false "Минимальный возраст (лет)"
// @Param age_to      query int    false "Максимальный возраст (лет)"
// @Param limit       query int    false "Лимит на страницу"
// @Param offset      query int    false "Смещение"
// @Security BearerAuth
// @Success 200 {array} models.Student
// @Failure 500 {object} helpers.ErrorResponse
// @Router /students [get]
func (h *StudentHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	_, claims, _ := jwtauth.FromContext(r.Context())
	role := claims["role"].(string)
	userID := int(claims["user_id"].(float64))

	q := r.URL.Query()

	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	var classIDPtr *int
	if v := q.Get("class_id"); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			classIDPtr = &id
		}
	}

	var gradeFromPtr, gradeToPtr *int
	if v := q.Get("grade_from"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			gradeFromPtr = &n
		}
	}
	if v := q.Get("grade_to"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			gradeToPtr = &n
		}
	}
	var ageFromPtr, ageToPtr *int
	if v := q.Get("age_from"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			ageFromPtr = &n
		}
	}
	if v := q.Get("age_to"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			ageToPtr = &n
		}
	}

	f := repository.StudentFilter{
		FullName:  q.Get("full_name"),
		Gender:    q.Get("gender"),
		ClassID:   classIDPtr,
		GradeFrom: gradeFromPtr,
		GradeTo:   gradeToPtr,
		AgeFrom:   ageFromPtr,
		AgeTo:     ageToPtr,
		Limit:     limit,
		Offset:    offset,
	}

	list, err := h.svc.GetAll(ctx, role, userID, f)
	if err != nil {
		if err.Error() == "access denied" {
			helpers.Error(w, http.StatusForbidden, "access denied")
			return
		}
		helpers.Error(w, http.StatusInternalServerError, "failed to get students")
		return
	}
	helpers.JSON(w, http.StatusOK, list)
}

// Create godoc
// @Summary Добавить ученика
// @Tags Students
// @Accept json
// @Produce json
// @Param data body models.Student true "Данные ученика"
// @Security BearerAuth
// @Success 201 {object} models.Student
// @Failure 400 {object} helpers.ErrorResponse
// @Failure 403 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Router /students [post]
func (h *StudentHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	_, claims, _ := jwtauth.FromContext(r.Context())
	role := claims["role"].(string)
	userID := int(claims["user_id"].(float64))

	var s models.Student
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		helpers.Error(w, http.StatusBadRequest, "invalid request")
		return
	}

	if s.FullName == "" || s.ClassID == 0 {
		helpers.Error(w, http.StatusBadRequest, "full_name and class_id required")
		return
	}

	if err := h.svc.Create(ctx, &s, role, userID); err != nil {
		if err.Error() == "access denied" {
			helpers.Error(w, http.StatusForbidden, "only schools can add students")
			return
		}
		helpers.Error(w, http.StatusInternalServerError, "failed to create student")
		return
	}
	helpers.JSON(w, http.StatusCreated, s)
}

// Delete godoc
// @Summary Удалить ученика
// @Tags Students
// @Param id path int true "ID ученика"
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 403 {object} helpers.ErrorResponse
// @Failure 404 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Router /students/{id} [delete]
func (h *StudentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	_, claims, _ := jwtauth.FromContext(r.Context())
	role := claims["role"].(string)
	userID := int(claims["user_id"].(float64))

	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	ok, err := h.svc.Delete(ctx, id, role, userID)
	if err != nil {
		if err.Error() == "access denied" {
			helpers.Error(w, http.StatusForbidden, "access denied")
			return
		}
		helpers.Error(w, http.StatusInternalServerError, "failed to delete student")
		return
	}
	if !ok {
		helpers.Error(w, http.StatusNotFound, "student not found")
		return
	}
	helpers.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ImportExcel godoc
// @Summary Импорт учеников из Excel
// @Description School — только в свою школу; ROO тоже можно использовать.
// @Tags Students
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Excel файл (.xlsx)"
// @Param strict query bool false "Строгий режим: если есть ошибки — никого не импортировать"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} helpers.ErrorResponse
// @Failure 403 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Router /students/import [post]
func (h *StudentHandler) ImportExcel(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	_, claims, _ := jwtauth.FromContext(r.Context())
	role := claims["role"].(string)
	userID := int(claims["user_id"].(float64))

	if role != "roo" && role != "school" {
		helpers.Error(w, http.StatusForbidden, "access denied")
		return
	}

	strict := r.URL.Query().Get("strict") == "1"

	file, header, err := r.FormFile("file")
	if err != nil {
		helpers.Error(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	f, err := excelize.OpenReader(file)
	if err != nil {
		helpers.Error(w, http.StatusBadRequest, "invalid excel file")
		return
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		helpers.Error(w, http.StatusBadRequest, "empty excel")
		return
	}
	rows, err := f.GetRows(sheetName)
	if err != nil {
		helpers.Error(w, http.StatusBadRequest, "failed to read rows")
		return
	}
	if len(rows) < 2 {
		helpers.Error(w, http.StatusBadRequest, "file has no data")
		return
	}

	headerRow := rows[0]
	colIndex := map[string]int{}
	for i, col := range headerRow {
		colIndex[col] = i
	}

	required := []string{"full_name", "class_id"}
	for _, col := range required {
		if _, ok := colIndex[col]; !ok {
			helpers.Error(w, http.StatusBadRequest, "missing column: "+col)
			return
		}
	}

	type RowError struct {
		Row   int    `json:"row"`
		Error string `json:"error"`
	}
	type rowData struct {
		rowNum  int
		student models.Student
	}

	var (
		parsedRows []rowData
		errorsList []RowError
	)

	// === 1. Парсинг и валидация ===
	for i, row := range rows[1:] {
		lineNum := i + 2

		get := func(name string) string {
			idx, ok := colIndex[name]
			if !ok || idx >= len(row) {
				return ""
			}
			return row[idx]
		}

		st := models.Student{
			FullName: get("full_name"),
			Phone:    strPtr(get("phone")),
			Address:  strPtr(get("address")),
			Note:     strPtr(get("note")),
		}

		if v := get("gender"); v != "" {
			g := v
			st.Gender = &g
		}
		if v := get("birth_date"); v != "" {
			if t, err := time.Parse("2006-01-02", v); err == nil {
				st.BirthDate = &t
			} else {
				errorsList = append(errorsList, RowError{Row: lineNum, Error: "invalid birth_date, expected YYYY-MM-DD"})
				continue
			}
		}
		if v := get("class_id"); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				st.ClassID = n
			} else {
				errorsList = append(errorsList, RowError{Row: lineNum, Error: "invalid class_id"})
				continue
			}
		}

		if st.FullName == "" || st.ClassID == 0 {
			errorsList = append(errorsList, RowError{Row: lineNum, Error: "full_name and class_id required"})
			continue
		}

		parsedRows = append(parsedRows, rowData{rowNum: lineNum, student: st})
	}

	if strict && len(errorsList) > 0 {
		helpers.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"file":        header.Filename,
			"created":     0,
			"errorsCount": len(errorsList),
			"errors":      errorsList,
			"strict":      true,
		})
		return
	}

	// === 2. Создание через сервис ===
	created := 0
	for _, item := range parsedRows {
		if err := h.svc.Create(ctx, &item.student, role, userID); err != nil {
			if strict {
				errorsList = append(errorsList, RowError{Row: item.rowNum, Error: err.Error()})
				break
			}
			errorsList = append(errorsList, RowError{Row: item.rowNum, Error: err.Error()})
			continue
		}
		created++
	}

	status := http.StatusOK
	if strict && created == 0 && len(errorsList) > 0 {
		status = http.StatusBadRequest
	}

	helpers.JSON(w, status, map[string]interface{}{
		"file":        header.Filename,
		"created":     created,
		"errorsCount": len(errorsList),
		"errors":      errorsList,
		"strict":      strict,
	})
}

// ImportTemplate godoc
// @Summary Скачать шаблон Excel для импорта учеников
// @Tags Students
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security BearerAuth
// @Success 200 {file} binary "Excel файл"
// @Router /students/import/template [get]
func (h *StudentHandler) ImportTemplate(w http.ResponseWriter, r *http.Request) {
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)

	headers := []string{
		"full_name",
		"birth_date", // YYYY-MM-DD
		"gender",     // male/female
		"phone",
		"address",
		"note",
		"class_id",
	}
	for i, hname := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, hname)
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=students_import_template.xlsx")

	if err := f.Write(w); err != nil {
		helpers.Error(w, http.StatusInternalServerError, "failed to write template")
		return
	}
}

// GetGradeStats godoc
// @Summary Получить статистику по диапазону классов
// @Description Возвращает количество классов и учеников в заданном диапазоне классов.
// @Tags Students
// @Produce json
// @Param grade_from query int false "Нижняя граница класса (номер)"
// @Param grade_to query int false "Верхняя граница класса (номер)"
// @Security BearerAuth
// @Success 200 {object} map[string]int
// @Failure 403 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Router /students/grade-stats [get]
func (h *StudentHandler) GetGradeStats(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	_, claims, _ := jwtauth.FromContext(r.Context())
	role := claims["role"].(string)
	userID := int(claims["user_id"].(float64))

	q := r.URL.Query()

	var gradeFromPtr, gradeToPtr *int
	if v := q.Get("grade_from"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			gradeFromPtr = &n
		}
	}
	if v := q.Get("grade_to"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			gradeToPtr = &n
		}
	}

	classesCount, studentsCount, err := h.svc.GetGradeStats(ctx, role, userID, gradeFromPtr, gradeToPtr)
	if err != nil {
		if err.Error() == "access denied" {
			helpers.Error(w, http.StatusForbidden, "access denied")
			return
		}
		helpers.Error(w, http.StatusInternalServerError, "failed to get grade stats")
		return
	}

	helpers.JSON(w, http.StatusOK, map[string]int{
		"classes":  classesCount,
		"students": studentsCount,
	})
}

// GetAgeStats godoc
// @Summary Количество детей по возрастам
// @Description Возвращает список {age, count} в заданном диапазоне возрастов (в годах).
// @Tags Students
// @Produce json
// @Param age_from query int false "Нижний возраст (лет)"
// @Param age_to query int false "Верхний возраст (лет)"
// @Security BearerAuth
// @Success 200 {array} models.AgeStat
// @Failure 403 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Router /students/age-stats [get]
func (h *StudentHandler) GetAgeStats(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	_, claims, _ := jwtauth.FromContext(r.Context())
	role := claims["role"].(string)
	userID := int(claims["user_id"].(float64))

	q := r.URL.Query()

	var ageFromPtr, ageToPtr *int
	if v := q.Get("age_from"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			ageFromPtr = &n
		}
	}
	if v := q.Get("age_to"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			ageToPtr = &n
		}
	}

	stats, err := h.svc.GetAgeStats(ctx, role, userID, ageFromPtr, ageToPtr)
	if err != nil {
		if err.Error() == "access denied" {
			helpers.Error(w, http.StatusForbidden, "access denied")
			return
		}
		helpers.Error(w, http.StatusInternalServerError, "failed to get age stats")
		return
	}

	helpers.JSON(w, http.StatusOK, stats)
}
