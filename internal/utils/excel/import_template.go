package excel

import "github.com/xuri/excelize/v2"

// BuildImportTemplateWorkbook создаёт шаблон с листом Students (на русском) и контактами на этом же листе.
func BuildImportTemplateWorkbook() (*excelize.File, error) {
	f := excelize.NewFile()
	main := "Students"
	_ = f.SetSheetName(f.GetSheetName(0), main)

	// Порядок колонок синхронизирован с парсером.
	headers := []string{
		"Номер дела", "Фамилия", "Имя", "Отчество",
		"Дата рождения (ГГГГ-ММ-ДД)", "Пол (м|ж)", "Гражданство",
		"ID школы", "Класс", "Год поступления",
		"Статус (обучается|переведён|выпущен|исключён)",
		"Адрес регистрации", "Адрес проживания", "Телефон ученика", "Email ученика",
		"СНИЛС", "Серия паспорта (4 цифры)", "Номер паспорта (6 цифр)", "Свидетельство о рождении",
		"Льготы", "Мед. примечания", "Группа здоровья (1-5)", "Аллергии", "Активности",
		"Согласие ПДн (да|нет)", "Дата ПДн (ГГГГ-ММ-ДД)",
		"Согласие Фото (да|нет)", "Дата Фото (ГГГГ-ММ-ДД)",
		"Согласие Интернет (да|нет)", "Дата Интернет (ГГГГ-ММ-ДД)",
		// Контакты (до трёх)
		"Контакт1 ФИО", "Контакт1 Телефон", "Контакт1 Связь",
		"Контакт2 ФИО", "Контакт2 Телефон", "Контакт2 Связь",
		"Контакт3 ФИО", "Контакт3 Телефон", "Контакт3 Связь",
	}
	for i, h := range headers {
		addr, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellStr(main, addr, h)
	}
	// 39 колонок => AM
	_ = f.AutoFilter(main, "A1:AM1", nil)
	_ = f.SetPanes(main, &excelize.Panes{Freeze: true, Split: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})

	// немного ширин
	_ = f.SetColWidth(main, "A", "C", 14)
	_ = f.SetColWidth(main, "D", "D", 14)
	_ = f.SetColWidth(main, "E", "G", 18)
	_ = f.SetColWidth(main, "H", "L", 14)
	_ = f.SetColWidth(main, "M", "P", 22)
	_ = f.SetColWidth(main, "Q", "T", 18)
	_ = f.SetColWidth(main, "U", "Y", 18)
	_ = f.SetColWidth(main, "Z", "AE", 18)
	_ = f.SetColWidth(main, "AF", "AM", 20)

	return f, nil
}
