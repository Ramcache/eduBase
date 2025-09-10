package templates

import "embed"

//go:embed students_import_template.xlsx
var FS embed.FS

func Bytes() ([]byte, error) {
	return FS.ReadFile("students_import_template.xlsx")
}
