package excel

import (
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"

	"eduBase/internal/models"
)

func BuildStudentsWorkbookFull(
	students []models.StudentCore,
	docs map[int]*models.StudentDocuments,
	medical map[int]*models.StudentMedical,
	consents map[int]*models.StudentConsents,
	contacts map[int][]models.EmergencyContact,
) (*excelize.File, error) {
	f := excelize.NewFile()

	// --- Sheet 1: Students (все основные поля в одну строку) ---
	main := "Students"
	_ = f.SetSheetName(f.GetSheetName(0), main)

	header := []string{
		"ID", "Номер дела",
		"Фамилия", "Имя", "Отчество",
		"Дата рождения", "Пол", "Гражданство",
		"ШколаID", "Класс", "Год поступления", "Статус",
		"Адрес регистрации", "Адрес проживания",
		"Телефон ученика", "Email ученика",
		// documents
		"СНИЛС", "Паспорт серия", "Паспорт номер", "Свидетельство о рождении",
		// medical
		"Льготы", "Мед. примечания", "Группа здоровья", "Аллергии", "Активности",
		// consents
		"Согласие ПДн", "Дата ПДн",
		"Согласие Фото", "Дата Фото",
		"Согласие Интернет", "Дата Интернет",
		// meta
		"Создано", "Обновлено",
	}
	for i, h := range header {
		cellAddr, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellStr(main, cellAddr, h)
	}

	for r, st := range students {
		row := r + 2
		_ = f.SetCellInt(main, cell(row, 1), int64(st.ID))
		_ = f.SetCellStr(main, cell(row, 2), st.StudentNumber)
		_ = f.SetCellStr(main, cell(row, 3), st.LastName)
		_ = f.SetCellStr(main, cell(row, 4), st.FirstName)
		_ = f.SetCellStr(main, cell(row, 5), strPtr(st.MiddleName))
		_ = f.SetCellStr(main, cell(row, 6), st.BirthDate.Format("02.01.2006"))
		_ = f.SetCellStr(main, cell(row, 7), string(st.Gender))
		_ = f.SetCellStr(main, cell(row, 8), strPtr(st.Citizenship))
		_ = f.SetCellInt(main, cell(row, 9), int64(st.SchoolID))
		_ = f.SetCellStr(main, cell(row, 10), st.ClassLabel)
		_ = f.SetCellInt(main, cell(row, 11), int64(st.AdmissionYear))
		_ = f.SetCellStr(main, cell(row, 12), string(st.Status))
		_ = f.SetCellStr(main, cell(row, 13), st.RegAddress)
		_ = f.SetCellStr(main, cell(row, 14), st.FactAddress)
		_ = f.SetCellStr(main, cell(row, 15), strPtr(st.StudentPhone))
		_ = f.SetCellStr(main, cell(row, 16), strPtr(st.StudentEmail))

		// documents
		if d := docs[st.ID]; d != nil {
			_ = f.SetCellStr(main, cell(row, 17), d.SNILS)
			_ = f.SetCellStr(main, cell(row, 18), strPtr(d.PassportSeries))
			_ = f.SetCellStr(main, cell(row, 19), strPtr(d.PassportNumber))
			_ = f.SetCellStr(main, cell(row, 20), strPtr(d.BirthCertificate))
		} else {
			_ = f.SetCellStr(main, cell(row, 17), "")
			_ = f.SetCellStr(main, cell(row, 18), "")
			_ = f.SetCellStr(main, cell(row, 19), "")
			_ = f.SetCellStr(main, cell(row, 20), "")
		}

		// medical
		if m := medical[st.ID]; m != nil {
			_ = f.SetCellStr(main, cell(row, 21), strPtr(m.Benefits))
			_ = f.SetCellStr(main, cell(row, 22), strPtr(m.MedicalNotes))
			_ = f.SetCellStr(main, cell(row, 23), intPtrToStr(m.HealthGroup))
			_ = f.SetCellStr(main, cell(row, 24), strPtr(m.Allergies))
			_ = f.SetCellStr(main, cell(row, 25), strPtr(m.Activities))
		} else {
			for c := 21; c <= 25; c++ {
				_ = f.SetCellStr(main, cell(row, c), "")
			}
		}

		// consents
		if c := consents[st.ID]; c != nil {
			_ = f.SetCellStr(main, cell(row, 26), boolToRU(c.ConsentPD))
			_ = f.SetCellStr(main, cell(row, 27), timePtr(c.ConsentPDDate))
			_ = f.SetCellStr(main, cell(row, 28), boolToRU(c.ConsentPhoto))
			_ = f.SetCellStr(main, cell(row, 29), timePtr(c.ConsentPhotoDate))
			_ = f.SetCellStr(main, cell(row, 30), boolToRU(c.ConsentInternet))
			_ = f.SetCellStr(main, cell(row, 31), timePtr(c.ConsentInternetDate))
		} else {
			for c := 26; c <= 31; c++ {
				_ = f.SetCellStr(main, cell(row, c), "")
			}
		}

		_ = f.SetCellStr(main, cell(row, 32), st.CreatedAt.Format(time.RFC3339))
		_ = f.SetCellStr(main, cell(row, 33), st.UpdatedAt.Format(time.RFC3339))
	}

	_ = f.AutoFilter(main, "A1:AG1", nil) // AG = 33 колонок
	_ = f.SetPanes(main, &excelize.Panes{Freeze: true, Split: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})
	_ = f.SetColWidth(main, "A", "A", 8)
	_ = f.SetColWidth(main, "B", "B", 14)
	_ = f.SetColWidth(main, "C", "E", 16)
	_ = f.SetColWidth(main, "F", "H", 13)
	_ = f.SetColWidth(main, "I", "L", 13)
	_ = f.SetColWidth(main, "M", "P", 22)
	_ = f.SetColWidth(main, "Q", "T", 18)
	_ = f.SetColWidth(main, "U", "Y", 18)
	_ = f.SetColWidth(main, "Z", "AE", 18)
	_ = f.SetColWidth(main, "AF", "AG", 22)

	// --- Sheet 2: Contacts (one-to-many) ---
	_, _ = f.NewSheet("Contacts")
	cont := "Contacts"
	_ = f.SetCellStr(cont, "A1", "StudentID")
	_ = f.SetCellStr(cont, "B1", "ФИО ученика")
	_ = f.SetCellStr(cont, "C1", "ФИО контакта")
	_ = f.SetCellStr(cont, "D1", "Телефон")
	_ = f.SetCellStr(cont, "E1", "Связь")
	_ = f.SetCellStr(cont, "F1", "Добавлен")

	row := 2
	for _, st := range students {
		fullName := st.LastName + " " + st.FirstName
		if st.MiddleName != nil && *st.MiddleName != "" {
			fullName += " " + *st.MiddleName
		}
		if arr := contacts[st.ID]; len(arr) > 0 {
			for _, c := range arr {
				_ = f.SetCellInt(cont, cell(row, 1), int64(st.ID))
				_ = f.SetCellStr(cont, cell(row, 2), fullName)
				_ = f.SetCellStr(cont, cell(row, 3), c.FullName)
				_ = f.SetCellStr(cont, cell(row, 4), c.Phone)
				_ = f.SetCellStr(cont, cell(row, 5), c.Relation)
				_ = f.SetCellStr(cont, cell(row, 6), c.CreatedAt.Format(time.RFC3339))
				row++
			}
		} else {
			_ = f.SetCellInt(cont, cell(row, 1), int64(st.ID))
			_ = f.SetCellStr(cont, cell(row, 2), fullName)
			_ = f.SetCellStr(cont, cell(row, 3), "")
			_ = f.SetCellStr(cont, cell(row, 4), "")
			_ = f.SetCellStr(cont, cell(row, 5), "")
			_ = f.SetCellStr(cont, cell(row, 6), "")
			row++
		}
	}
	_ = f.AutoFilter(cont, "A1:F1", nil)
	_ = f.SetPanes(cont, &excelize.Panes{Freeze: true, Split: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})
	_ = f.SetColWidth(cont, "A", "A", 10)
	_ = f.SetColWidth(cont, "B", "C", 28)
	_ = f.SetColWidth(cont, "D", "D", 18)
	_ = f.SetColWidth(cont, "E", "F", 16)

	return f, nil
}

func strPtr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
func intPtrToStr(p *int) string {
	if p == nil {
		return ""
	}
	return strconv.Itoa(*p)
}
func timePtr(p *time.Time) string {
	if p == nil || p.IsZero() {
		return ""
	}
	return p.Format("2006-01-02")
}
func boolToRU(b bool) string {
	if b {
		return "да"
	}
	return "нет"
}
