package excel

import (
	"errors"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"

	"eduBase/internal/service"
)

func ParseImportWorkbook(f *excelize.File) ([]service.ImportStudent, []service.ImportContact, error) {
	// Сначала читаем студентов + контакты из этого же листа
	students, contacts, err := parseStudentsSheetWithInlineContacts(f)
	if err != nil {
		return nil, nil, err
	}

	// Back-compat: если вдруг присутствует лист Contacts/Контакты — тоже добавим
	if extra, err2 := parseContactsSheetOptional(f); err2 == nil && len(extra) > 0 {
		contacts = append(contacts, extra...)
	}
	return students, contacts, nil
}

// ===== Students sheet with inline contacts =====

func parseStudentsSheetWithInlineContacts(f *excelize.File) ([]service.ImportStudent, []service.ImportContact, error) {
	const sheet = "Students"
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, nil, errors.New("лист 'Students' не найден")
	}
	if len(rows) < 2 {
		return []service.ImportStudent{}, []service.ImportContact{}, nil
	}

	out := make([]service.ImportStudent, 0, len(rows)-1)
	contacts := make([]service.ImportContact, 0, len(rows)-1)

	// ожидаем 39 колонок (30 базовых + 9 контактов)
	for i := 1; i < len(rows); i++ {
		r := padRow(rows[i], 39)
		// пропуск пустых строк
		if strings.TrimSpace(r[0]) == "" && strings.TrimSpace(r[1]) == "" {
			continue
		}

		item := service.ImportStudent{
			StudentNumber:       strings.TrimSpace(r[0]),
			LastName:            strings.TrimSpace(r[1]),
			FirstName:           strings.TrimSpace(r[2]),
			MiddleName:          optStr(r[3]),
			BirthDate:           strings.TrimSpace(r[4]),
			Gender:              normalizeGender(r[5]),
			Citizenship:         optStr(r[6]),
			SchoolID:            mustAtoi(r[7]),
			ClassLabel:          strings.TrimSpace(r[8]),
			AdmissionYear:       mustAtoi(r[9]),
			Status:              normalizeStatus(r[10]),
			RegAddress:          strings.TrimSpace(r[11]),
			FactAddress:         strings.TrimSpace(r[12]),
			StudentPhone:        optStr(r[13]),
			StudentEmail:        optStr(r[14]),
			SNILS:               optStr(r[15]),
			PassportSeries:      optStr(r[16]),
			PassportNumber:      optStr(r[17]),
			BirthCertificate:    optStr(r[18]),
			Benefits:            optStr(r[19]),
			MedicalNotes:        optStr(r[20]),
			HealthGroup:         optInt(r[21]),
			Allergies:           optStr(r[22]),
			Activities:          optStr(r[23]),
			ConsentPD:           optBool(r[24]),
			ConsentPDDate:       optStr(r[25]),
			ConsentPhoto:        optBool(r[26]),
			ConsentPhotoDate:    optStr(r[27]),
			ConsentInternet:     optBool(r[28]),
			ConsentInternetDate: optStr(r[29]),
		}
		out = append(out, item)

		// контакты из этих же колонок
		sn := item.StudentNumber
		addContact(&contacts, sn, r[30], r[31], r[32])
		addContact(&contacts, sn, r[33], r[34], r[35])
		addContact(&contacts, sn, r[36], r[37], r[38])
	}

	return out, contacts, nil
}

func addContact(dst *[]service.ImportContact, sn, full, phone, rel string) {
	full = strings.TrimSpace(full)
	phone = strings.TrimSpace(phone)
	rel = strings.TrimSpace(rel)
	// добавляем, если есть хоть что-то
	if full == "" && phone == "" && rel == "" {
		return
	}
	*dst = append(*dst, service.ImportContact{
		StudentNumber: sn,
		FullName:      full,
		Phone:         phone,
		Relation:      rel,
	})
}

// ===== Optional Contacts sheet (back-compat) =====

func parseContactsSheetOptional(f *excelize.File) ([]service.ImportContact, error) {
	name := findContactsSheet(f)
	if name == "" {
		return []service.ImportContact{}, nil
	}
	rows, err := f.GetRows(name)
	if err != nil || len(rows) < 2 {
		return []service.ImportContact{}, nil
	}
	out := make([]service.ImportContact, 0, len(rows)-1)
	for i := 1; i < len(rows); i++ {
		r := padRow(rows[i], 4)
		if strings.TrimSpace(r[0]) == "" && strings.TrimSpace(r[1]) == "" {
			continue
		}
		out = append(out, service.ImportContact{
			StudentNumber: strings.TrimSpace(r[0]),
			FullName:      strings.TrimSpace(r[1]),
			Phone:         strings.TrimSpace(r[2]),
			Relation:      strings.TrimSpace(r[3]),
		})
	}
	return out, nil
}

func findContactsSheet(f *excelize.File) string {
	for _, name := range f.GetSheetList() {
		n := strings.ToLower(strings.TrimSpace(name))
		if n == "contacts" || n == "контакты" {
			return name
		}
	}
	// эвристика по шапке
	for _, name := range f.GetSheetList() {
		rows, err := f.GetRows(name)
		if err != nil || len(rows) == 0 {
			continue
		}
		h := padRow(rows[0], 4)
		if headerLooksLikeContacts(h) {
			return name
		}
	}
	return ""
}

func headerLooksLikeContacts(h []string) bool {
	if len(h) < 4 {
		return false
	}
	a := norm(h[0])
	b := norm(h[1])
	c := norm(h[2])
	d := norm(h[3])
	okA := in(a, "student_number", "номер дела", "номер_дела")
	okB := in(b, "фио контакта", "contact_full_name", "фио")
	okC := in(c, "телефон", "phone")
	okD := in(d, "связь", "relation")
	return okA && okB && okC && okD
}

// ===== helpers =====

func padRow(row []string, cols int) []string {
	if len(row) >= cols {
		return row
	}
	out := make([]string, cols)
	copy(out, row)
	return out
}
func optStr(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}
func mustAtoi(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	v, _ := strconv.Atoi(s)
	return v
}
func optInt(s string) *int {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &v
}
func optBool(s string) *bool {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return nil
	}
	switch s {
	case "true", "1", "t", "y", "yes", "да", "д", "истина":
		b := true
		return &b
	case "false", "0", "f", "n", "no", "нет", "н", "ложь":
		b := false
		return &b
	default:
		return nil
	}
}
func normalizeGender(s string) string {
	x := strings.TrimSpace(strings.ToLower(s))
	switch x {
	case "м", "муж", "мужской", "male", "m":
		return "m"
	case "ж", "жен", "женский", "female", "f":
		return "f"
	default:
		return s
	}
}
func normalizeStatus(s string) string {
	x := strings.TrimSpace(strings.ToLower(s))
	switch x {
	case "обучается", "зачислен", "зачислена", "учится":
		return "enrolled"
	case "переведён", "переведен", "переведена":
		return "transferred"
	case "выпущен", "выпущена", "окончил", "окончила", "выпускник":
		return "graduated"
	case "исключён", "исключен", "исключена":
		return "expelled"
	default:
		return s
	}
}
func norm(s string) string { return strings.TrimSpace(strings.ToLower(s)) }
func in(s string, opts ...string) bool {
	for _, o := range opts {
		if s == o {
			return true
		}
	}
	return false
}
