package models

type StatsSummary struct {
	Schools    int `json:"schools"`
	Classes    int `json:"classes"`
	Students   int `json:"students"`
	Teachers   int `json:"teachers"`
	StaffTotal int `json:"staff_total"`
}
