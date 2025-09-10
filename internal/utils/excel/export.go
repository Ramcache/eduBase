package excel

import (
	"time"

	"github.com/xuri/excelize/v2"

	"eduBase/internal/models"
)

func BuildStudentsWorkbook(items []models.StudentListItem) (*excelize.File, error) {
	f := excelize.NewFile()
	sheet := "Students"
	_ = f.SetSheetName(f.GetSheetName(0), sheet)

	// Заголовки
	header := []string{
		"ID", "Номер дела", "ФИО", "Дата рождения", "Пол",
		"ШколаID", "Класс", "Год поступления", "Статус", "Создано",
	}
	for i, h := range header {
		cellAddr, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellStr(sheet, cellAddr, h)
	}

	// Данные
	for r, it := range items {
		row := r + 2
		_ = f.SetCellInt(sheet, cell(row, 1), int64(it.ID)) // int
		_ = f.SetCellStr(sheet, cell(row, 2), it.StudentNumber)
		_ = f.SetCellStr(sheet, cell(row, 3), it.FullName)
		_ = f.SetCellStr(sheet, cell(row, 4), it.BirthDate.Format("2006-01-02"))
		_ = f.SetCellStr(sheet, cell(row, 5), string(it.Gender))
		_ = f.SetCellInt(sheet, cell(row, 6), int64(it.SchoolID)) // int
		_ = f.SetCellStr(sheet, cell(row, 7), it.ClassLabel)
		_ = f.SetCellInt(sheet, cell(row, 8), int64(it.AdmissionYear)) // int
		_ = f.SetCellStr(sheet, cell(row, 9), string(it.Status))
		_ = f.SetCellStr(sheet, cell(row, 10), it.CreatedAt.Format(time.RFC3339))
	}

	// Фильтр по заголовкам: новый сигнатурный формат AutoFilter
	_ = f.AutoFilter(sheet, "A1:J1", nil)

	// Приклеить шапку
	_ = f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		Split:       true,
		XSplit:      0,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	})

	// Ширины колонок
	_ = f.SetColWidth(sheet, "A", "A", 8)
	_ = f.SetColWidth(sheet, "B", "B", 14)
	_ = f.SetColWidth(sheet, "C", "C", 28)
	_ = f.SetColWidth(sheet, "D", "E", 14)
	_ = f.SetColWidth(sheet, "F", "H", 12)
	_ = f.SetColWidth(sheet, "I", "J", 18)

	return f, nil
}

func cell(row, col int) string {
	c, _ := excelize.CoordinatesToCellName(col, row)
	return c
}
