package handlers

import (
	"context"
	"encoding/json"
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

type StaffHandler struct {
	svc *services.StaffService
}

func NewStaffHandler(svc *services.StaffService) *StaffHandler {
	return &StaffHandler{svc: svc}
}

func (h *StaffHandler) Routes(r chi.Router) {
	r.Route("/staff", func(r chi.Router) {
		r.Get("/", h.GetAll)
		r.Get("/{id}", h.GetByID)
		r.Get("/stats", h.GetStats)
		r.Get("/import/template", h.ImportTemplate)
		r.Post("/", h.Create)
		r.Post("/import", h.ImportExcel)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})
}

// GetByID godoc
// @Summary Получить сотрудника по ID
// @Description ROO — любого, School — только своего
// @Tags Staff
// @Produce json
// @Param id path int true "ID сотрудника"
// @Security BearerAuth
// @Success 200 {object} models.Staff
// @Failure 403 {object} helpers.ErrorResponse
// @Failure 404 {object} helpers.ErrorResponse
// @Router /staff/{id} [get]
func (h *StaffHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	_, claims, _ := jwtauth.FromContext(r.Context())
	role := claims["role"].(string)
	userID := int(claims["user_id"].(float64))

	staff, err := h.svc.GetByID(ctx, id, role, userID)
	if err != nil {
		if err.Error() == "access denied" {
			helpers.Error(w, http.StatusForbidden, "access denied")
			return
		}
		helpers.Error(w, http.StatusNotFound, "staff not found")
		return
	}

	helpers.JSON(w, http.StatusOK, staff)
}

// Update godoc
// @Summary Обновить данные сотрудника
// @Description ROO — любого, School — только своего
// @Tags Staff
// @Accept json
// @Produce json
// @Param id path int true "ID сотрудника"
// @Param data body models.Staff true "Обновлённые данные"
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 403 {object} helpers.ErrorResponse
// @Failure 404 {object} helpers.ErrorResponse
// @Router /staff/{id} [put]
func (h *StaffHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	_, claims, _ := jwtauth.FromContext(r.Context())
	role := claims["role"].(string)
	userID := int(claims["user_id"].(float64))

	var s models.Staff
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		helpers.Error(w, http.StatusBadRequest, "invalid request")
		return
	}

	if s.FullName == "" || s.Phone == "" || s.Position == "" {
		helpers.Error(w, http.StatusBadRequest, "full_name, phone and position required")
		return
	}

	ok, err := h.svc.Update(ctx, id, &s, role, userID)
	if err != nil {
		if err.Error() == "access denied" {
			helpers.Error(w, http.StatusForbidden, "access denied")
			return
		}
		helpers.Error(w, http.StatusInternalServerError, "failed to update staff")
		return
	}
	if !ok {
		helpers.Error(w, http.StatusNotFound, "staff not found")
		return
	}

	helpers.JSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// GetStats godoc
// @Summary Получить статистику по персоналу (ROO)
// @Description Кол-во сотрудников по должностям
// @Tags Staff
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]int
// @Failure 403 {object} helpers.ErrorResponse
// @Router /staff/stats [get]
func (h *StaffHandler) GetStats(w http.ResponseWriter, r *http.Request) {
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

// GetAll godoc
// @Summary Получить список сотрудников
// @Description ROO — всех, School — только своих
// @Tags Staff
// @Produce json
// @Param full_name query string false "ФИО"
// @Param phone query string false "Телефон"
// @Param position query string false "Должность"
// @Param subject query string false "Предмет"
// @Param education query string false "Образование"
// @Param category query string false "Категория"
// @Param ped_experience query int false "Минимальный пед. стаж"
// @Param total_experience query int false "Минимальный общий стаж"
// @Param limit query int false "Лимит на страницу"
// @Param offset query int false "Смещение"
// @Security BearerAuth
// @Success 200 {array} models.Staff
// @Failure 500 {object} helpers.ErrorResponse
// @Router /staff [get]
func (h *StaffHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	_, claims, _ := jwtauth.FromContext(r.Context())
	role := claims["role"].(string)
	userID := int(claims["user_id"].(float64))

	q := r.URL.Query()

	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	var pedExpPtr, totalExpPtr *int
	if v := q.Get("ped_experience"); v != "" {
		if x, err := strconv.Atoi(v); err == nil {
			pedExpPtr = &x
		}
	}
	if v := q.Get("total_experience"); v != "" {
		if x, err := strconv.Atoi(v); err == nil {
			totalExpPtr = &x
		}
	}

	filter := repository.StaffFilter{
		FullName:        q.Get("full_name"),
		Phone:           q.Get("phone"),
		Position:        q.Get("position"),
		Subject:         q.Get("subject"),
		Education:       q.Get("education"),
		Category:        q.Get("category"),
		PedExperience:   pedExpPtr,
		TotalExperience: totalExpPtr,
		Limit:           limit,
		Offset:          offset,
	}

	list, err := h.svc.GetAll(ctx, role, userID, filter)
	if err != nil {
		if err.Error() == "access denied" {
			helpers.Error(w, http.StatusForbidden, "access denied")
			return
		}
		helpers.Error(w, http.StatusInternalServerError, "failed to get staff")
		return
	}

	helpers.JSON(w, http.StatusOK, list)
}

// Create godoc
// @Summary Добавить сотрудника
// @Description Только школа может добавлять своих сотрудников
// @Tags Staff
// @Accept json
// @Produce json
// @Param data body models.Staff true "Данные сотрудника"
// @Security BearerAuth
// @Success 201 {object} models.Staff
// @Failure 400 {object} helpers.ErrorResponse
// @Failure 403 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Router /staff [post]
func (h *StaffHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	_, claims, _ := jwtauth.FromContext(r.Context())
	role := claims["role"].(string)
	userID := int(claims["user_id"].(float64))

	if role != "school" {
		helpers.Error(w, http.StatusForbidden, "only schools can add staff")
		return
	}

	var s models.Staff
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		helpers.Error(w, http.StatusBadRequest, "invalid request")
		return
	}

	if s.FullName == "" || s.Phone == "" || s.Position == "" {
		helpers.Error(w, http.StatusBadRequest, "full_name, phone and position are required")
		return
	}

	if err := h.svc.Create(ctx, &s, role, userID); err != nil {
		if err.Error() == "access denied" {
			helpers.Error(w, http.StatusForbidden, "access denied")
			return
		}
		helpers.Error(w, http.StatusInternalServerError, "failed to create staff")
		return
	}

	helpers.JSON(w, http.StatusCreated, s)
}

// Delete godoc
// @Summary Удалить сотрудника
// @Description School — только своего, ROO — любого
// @Tags Staff
// @Param id path int true "ID сотрудника"
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 403 {object} helpers.ErrorResponse
// @Failure 404 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Router /staff/{id} [delete]
func (h *StaffHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
		helpers.Error(w, http.StatusInternalServerError, "failed to delete staff")
		return
	}
	if !ok {
		helpers.Error(w, http.StatusNotFound, "staff not found")
		return
	}

	helpers.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ImportExcel godoc
// @Summary Импорт сотрудников из Excel
// @Description School — только в свою школу; ROO — может импортировать для любой школы (по логике сервиса).
// @Tags Staff
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Excel файл (.xlsx)"
// @Param strict query bool false "Строгий режим: если есть ошибки — никого не импортировать"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} helpers.ErrorResponse
// @Failure 403 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Router /staff/import [post]
func (h *StaffHandler) ImportExcel(w http.ResponseWriter, r *http.Request) {
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

	required := []string{"full_name", "phone", "position"}
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
		rowNum int
		staff  models.Staff
	}

	var (
		parsedRows []rowData
		errorsList []RowError
	)

	// === 1. Парсинг и валидация ===
	for i, row := range rows[1:] {
		lineNum := i + 2 // человекочитаемый номер строки

		get := func(name string) string {
			idx, ok := colIndex[name]
			if !ok || idx >= len(row) {
				return ""
			}
			return row[idx]
		}

		st := models.Staff{
			FullName:  get("full_name"),
			Phone:     get("phone"),
			Position:  get("position"),
			Subject:   strPtr(get("subject")),
			Education: strPtr(get("education")),
			Category:  strPtr(get("category")),
			Note:      strPtr(get("note")),
		}

		if st.FullName == "" || st.Phone == "" || st.Position == "" {
			errorsList = append(errorsList, RowError{Row: lineNum, Error: "full_name, phone and position are required"})
			continue
		}

		if v := get("ped_experience"); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				st.PedExperience = &n
			} else {
				errorsList = append(errorsList, RowError{Row: lineNum, Error: "invalid ped_experience"})
				continue
			}
		}
		if v := get("total_experience"); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				st.TotalExperience = &n
			} else {
				errorsList = append(errorsList, RowError{Row: lineNum, Error: "invalid total_experience"})
				continue
			}
		}
		if v := get("work_start"); v != "" {
			if t, err := time.Parse("2006-01-02", v); err == nil {
				st.WorkStart = &t
			} else {
				errorsList = append(errorsList, RowError{Row: lineNum, Error: "invalid work_start, expected YYYY-MM-DD"})
				continue
			}
		}

		parsedRows = append(parsedRows, rowData{rowNum: lineNum, staff: st})
	}

	// В строгом режиме при любом parse-errore — никого не импортируем
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
		if err := h.svc.Create(ctx, &item.staff, role, userID); err != nil {
			if strict {
				errorsList = append(errorsList, RowError{Row: item.rowNum, Error: err.Error()})
				// в строгом режиме при первой же ошибке — выходим
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

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// ImportTemplate godoc
// @Summary Скачать шаблон Excel для импорта сотрудников
// @Tags Staff
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security BearerAuth
// @Success 200 {file} binary "Excel файл"
// @Router /staff/import/template [get]
func (h *StaffHandler) ImportTemplate(w http.ResponseWriter, r *http.Request) {
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)

	// шапка
	headers := []string{
		"full_name",
		"phone",
		"position",
		"subject",
		"education",
		"category",
		"ped_experience",
		"total_experience",
		"work_start", // формат YYYY-MM-DD
		"note",
	}
	for i, hname := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, hname)
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=staff_import_template.xlsx")

	if err := f.Write(w); err != nil {
		helpers.Error(w, http.StatusInternalServerError, "failed to write template")
		return
	}
}
