package models

type StatsSummary struct {
	Classes    int `json:"classes"`
	Students   int `json:"students"`
	Teachers   int `json:"teachers"`
	StaffTotal int `json:"staff_total"`
}
