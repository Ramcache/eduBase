package handlers

import (
	"eduBase/internal/repository"
	"eduBase/internal/service"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

type StudentsCoreHandler struct{ svc service.StudentService }

func NewStudentsCoreHandler(s service.StudentService) *StudentsCoreHandler {
	return &StudentsCoreHandler{svc: s}
}

type createCoreReq struct {
	StudentNumber string  `json:"student_number"`
	LastName      string  `json:"last_name"`
	FirstName     string  `json:"first_name"`
	MiddleName    *string `json:"middle_name"`
	BirthDate     string  `json:"birth_date"` // YYYY-MM-DD
	Gender        string  `json:"gender"`
	Citizenship   *string `json:"citizenship"`
	SchoolID      int     `json:"school_id"`
	ClassLabel    string  `json:"class_label"`
	AdmissionYear int     `json:"admission_year"`
	Status        string  `json:"status"`
	RegAddress    string  `json:"reg_address"`
	FactAddress   string  `json:"fact_address"`
	StudentPhone  *string `json:"student_phone"`
	StudentEmail  *string `json:"student_email"`
}

// Create
// @Summary      Create student core
// @Description  Создать базовую карточку ученика (А+Б+В)
// @Tags         Students
// @Accept       json
// @Produce      json
// @Param        payload  body      handlers.CreateStudentCoreRequest  true  "Core payload"
// @Success      201      {object}  handlers.CreateIDResponse
// @Failure      422      {string}  string  "validation error"
// @Router       /api/students [post]
func (h *StudentsCoreHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createCoreReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", 400)
		return
	}
	id, err := h.svc.CreateCore(r.Context(), service.CreateCoreDTO(req), userID(r))
	if err != nil {
		http.Error(w, err.Error(), 422)
		return
	}
	writeJSON(w, 201, map[string]any{"id": id})
}

// Update
// @Summary      Update student core (PATCH)
// @Description  Частичное/полное обновление базовой карточки (в текущей реализации как PUT)
// @Tags         Students
// @Accept       json
// @Produce      json
// @Param        id       path      int                                true  "Student ID"
// @Param        payload  body      handlers.UpdateStudentCoreRequest   true  "Core payload"
// @Success      200      {object}  handlers.OkResponse
// @Failure      422      {string}  string  "validation error"
// @Router       /api/students/{id} [patch]
func (h *StudentsCoreHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var req createCoreReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", 400)
		return
	}
	err := h.svc.UpdateCore(r.Context(), id, service.UpdateCoreDTO(req), userID(r))
	if err != nil {
		http.Error(w, err.Error(), 422)
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func (h *StudentsCoreHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	view, err := h.svc.GetAggregate(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", 404)
		return
	}
	writeJSON(w, 200, view)
}

// List
// @Summary      List students
// @Description  Список учеников с фильтрами и пагинацией
// @Tags         Students
// @Produce      json
// @Param        q                     query    string  false  "Поиск по ФИО (ILIKE)"
// @Param        school_id             query    int     false  "ID школы"
// @Param        class                 query    string  false  "Класс (например 7А)"
// @Param        status                query    string  false  "enrolled|transferred|graduated|expelled"
// @Param        admission_year_from   query    int     false  "Год поступления с"
// @Param        admission_year_to     query    int     false  "Год поступления по"
// @Param        birth_date_from       query    string  false  "Дата рождения c (YYYY-MM-DD)"
// @Param        birth_date_to         query    string  false  "Дата рождения по (YYYY-MM-DD)"
// @Param        limit                 query    int     false  "Лимит"    default(50)
// @Param        offset                query    int     false  "Смещение" default(0)
// @Success      200  {object}  handlers.StudentListResponse
// @Router       /api/students [get]
func (h *StudentsCoreHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	classLabel := r.URL.Query().Get("class")
	status := r.URL.Query().Get("status")
	schoolID := parseIntPtr(r.URL.Query().Get("school_id"))

	ayFrom := parseIntPtr(r.URL.Query().Get("admission_year_from"))
	ayTo := parseIntPtr(r.URL.Query().Get("admission_year_to"))
	bdFrom := parseDatePtr(r.URL.Query().Get("birth_date_from"))
	bdTo := parseDatePtr(r.URL.Query().Get("birth_date_to"))

	limit := atoiDefault(r.URL.Query().Get("limit"), 50)
	offset := atoiDefault(r.URL.Query().Get("offset"), 0)

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
	items, total, err := h.svc.List(r.Context(), f, limit, offset)
	if err != nil {
		http.Error(w, "db error", 500)
		return
	}
	writeJSON(w, 200, map[string]any{
		"total":  total,
		"limit":  limit,
		"offset": offset,
		"items":  items,
	})
}

func atoiDefault(s string, d int) int {
	if s == "" {
		return d
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return d
	}
	return n
}
