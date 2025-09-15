package service

import (
	"context"
	"eduBase/internal/validators"
	"errors"
	"strconv"
	"strings"
	"time"
)

type ImportStudent struct {
	// core
	StudentNumber string
	LastName      string
	FirstName     string
	MiddleName    *string
	BirthDate     string
	Gender        string
	Citizenship   *string
	SchoolID      int
	ClassLabel    string
	AdmissionYear int
	Status        string
	RegAddress    string
	FactAddress   string
	StudentPhone  *string
	StudentEmail  *string

	// documents
	SNILS            *string
	PassportSeries   *string
	PassportNumber   *string
	BirthCertificate *string

	// medical
	Benefits     *string
	MedicalNotes *string
	HealthGroup  *int
	Allergies    *string
	Activities   *string

	// consents
	ConsentPD           *bool
	ConsentPDDate       *string
	ConsentPhoto        *bool
	ConsentPhotoDate    *string
	ConsentInternet     *bool
	ConsentInternetDate *string
}

type ImportContact struct {
	StudentNumber string
	FullName      string
	Phone         string
	Relation      string
}

type ImportResult struct {
	Row           int     `json:"row"`
	StudentNumber string  `json:"student_number"`
	StudentID     *int    `json:"student_id,omitempty"`
	Created       bool    `json:"created"`
	Updated       bool    `json:"updated"`
	Error         *string `json:"error,omitempty"`
}

func (s *studentService) Import(ctx context.Context, rows []ImportStudent, contacts []ImportContact, operatorID int, replaceContacts bool) ([]ImportResult, error) {
	results := make([]ImportResult, 0, len(rows))

	// Индекс контактов по StudentNumber
	bySN := map[string][]ImportContact{}
	for _, c := range contacts {
		if c.StudentNumber == "" {
			continue
		}
		bySN[c.StudentNumber] = append(bySN[c.StudentNumber], c)
	}

	for i, r := range rows {
		res := ImportResult{Row: i + 2, StudentNumber: r.StudentNumber} // +2: шапка в A1
		if r.StudentNumber == "" {
			msg := "student_number обязателен"
			res.Error = &msg
			results = append(results, res)
			continue
		}
		// базовая валидация ядра
		if err := validators.BasicCore(r.LastName, r.FirstName, r.BirthDate, r.Gender, r.SchoolID, r.ClassLabel, r.AdmissionYear, r.RegAddress, r.FactAddress, r.Status); err != nil {
			msg := err.Error()
			res.Error = &msg
			results = append(results, res)
			continue
		}
		bd, err := parseDateLooseToTime(r.BirthDate)
		if err != nil {
			msg := "birth_date: ожидаю ДД.ММ.ГГГГ или YYYY-MM-DD (также принимаю DD-MM-YYYY и excel-число)"
			res.Error = &msg
			results = append(results, res)
			continue
		}

		// документы: SNILS обязателен всегда
		if r.SNILS == nil || *r.SNILS == "" {
			msg := "СНИЛС обязателен"
			res.Error = &msg
			results = append(results, res)
			continue
		}
		if err := validators.CheckDocs(bd, r.SNILS, r.PassportSeries, r.PassportNumber, r.BirthCertificate); err != nil {
			msg := err.Error()
			res.Error = &msg
			results = append(results, res)
			continue
		}

		// есть ли уже ученик с таким номером дела?
		existingID, ok, err := s.coreRepo.FindIDByStudentNumber(ctx, r.StudentNumber)
		if err != nil {
			msg := err.Error()
			res.Error = &msg
			results = append(results, res)
			continue
		}

		if !ok {
			newID, err := s.CreateCore(ctx, CreateCoreDTO{
				StudentNumber: r.StudentNumber,
				LastName:      r.LastName,
				FirstName:     r.FirstName,
				MiddleName:    r.MiddleName,
				BirthDate:     r.BirthDate,
				Gender:        r.Gender,
				Citizenship:   r.Citizenship,
				SchoolID:      r.SchoolID,
				ClassLabel:    r.ClassLabel,
				AdmissionYear: r.AdmissionYear,
				Status:        r.Status,
				RegAddress:    r.RegAddress,
				FactAddress:   r.FactAddress,
				StudentPhone:  r.StudentPhone,
				StudentEmail:  r.StudentEmail,
			}, operatorID)
			if err != nil {
				msg := err.Error()
				res.Error = &msg
				results = append(results, res)
				continue
			}
			existingID = newID
			res.StudentID = &existingID
			res.Created = true
		} else {
			err := s.UpdateCore(ctx, existingID, UpdateCoreDTO{
				StudentNumber: r.StudentNumber,
				LastName:      r.LastName,
				FirstName:     r.FirstName,
				MiddleName:    r.MiddleName,
				BirthDate:     r.BirthDate,
				Gender:        r.Gender,
				Citizenship:   r.Citizenship,
				SchoolID:      r.SchoolID,
				ClassLabel:    r.ClassLabel,
				AdmissionYear: r.AdmissionYear,
				Status:        r.Status,
				RegAddress:    r.RegAddress,
				FactAddress:   r.FactAddress,
				StudentPhone:  r.StudentPhone,
				StudentEmail:  r.StudentEmail,
			}, operatorID)
			if err != nil {
				msg := err.Error()
				res.Error = &msg
				results = append(results, res)
				continue
			}
			res.StudentID = &existingID
			res.Updated = true
		}

		// документы
		err = s.UpsertDocuments(ctx, DocumentsDTO{
			StudentID:        existingID,
			SNILS:            derefStr(r.SNILS),
			PassportSeries:   r.PassportSeries,
			PassportNumber:   r.PassportNumber,
			BirthCertificate: r.BirthCertificate,
			BirthDate:        r.BirthDate,
		})
		if err != nil {
			msg := "documents: " + err.Error()
			res.Error = &msg
			results = append(results, res)
			continue
		}

		// medical
		err = s.UpsertMedical(ctx, MedicalDTO{
			StudentID:    existingID,
			Benefits:     r.Benefits,
			MedicalNotes: r.MedicalNotes,
			HealthGroup:  r.HealthGroup,
			Allergies:    r.Allergies,
			Activities:   r.Activities,
		})
		if err != nil {
			msg := "medical: " + err.Error()
			res.Error = &msg
			results = append(results, res)
			continue
		}

		// consents
		cp := false
		if r.ConsentPD != nil {
			cp = *r.ConsentPD
		}
		cph := false
		if r.ConsentPhoto != nil {
			cph = *r.ConsentPhoto
		}
		ci := false
		if r.ConsentInternet != nil {
			ci = *r.ConsentInternet
		}
		err = s.UpsertConsents(ctx, ConsentsDTO{
			StudentID:           existingID,
			ConsentPD:           cp,
			ConsentPDDate:       r.ConsentPDDate,
			ConsentPhoto:        cph,
			ConsentPhotoDate:    r.ConsentPhotoDate,
			ConsentInternet:     ci,
			ConsentInternetDate: r.ConsentInternetDate,
		})
		if err != nil {
			msg := "consents: " + err.Error()
			res.Error = &msg
			results = append(results, res)
			continue
		}

		// contacts
		if list := bySN[r.StudentNumber]; len(list) > 0 {
			if replaceContacts {
				if err := s.contactsRepo.DeleteByStudent(ctx, existingID); err != nil {
					msg := "contacts delete: " + err.Error()
					res.Error = &msg
					results = append(results, res)
					continue
				}
			}
			for _, c := range list {
				if _, err := s.contactsRepo.Add(ctx, existingID, c.FullName, c.Phone, c.Relation); err != nil {
					msg := "contacts add: " + err.Error()
					res.Error = &msg
					results = append(results, res)
					continue
				}
			}
		}

		results = append(results, res)
	}
	return results, nil
}

func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// parseDateLooseToTime — понимает DD.MM.YYYY / YYYY-MM-DD / DD-MM-YYYY,
// варианты с временем, "г.", Excel serial number и однозначные день/месяц.
func parseDateLooseToTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, errors.New("empty date")
	}

	// Excel serial number (пример: 45565)
	if isAllDigits(s) {
		if v, err := strconv.Atoi(s); err == nil && v > 20000 && v < 100000 {
			base := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
			return base.AddDate(0, 0, v), nil
		}
	}

	// Срезаем время/зоны/лишние пометки: " 00:00:00", "T00:00:00Z", " г.", "г."
	s = stripDecor(s)

	// Пробуем набор layout'ов (принимаем и 1-значные день/месяц)
	layouts := []string{
		"02.01.2006", "2.1.2006", "02.01.06", "2.1.06",
		"2006-01-02", "2006-1-2",
		"02-01-2006", "2-1-2006", "02-01-06", "2-1-06",
		// частые варианты с временем
		"02.01.2006 15:04", "2.1.2006 15:04", "02.01.2006 15:04:05", "2.1.2006 15:04:05",
		"2006-01-02 15:04", "2006-1-2 15:04", "2006-01-02 15:04:05", "2006-1-2 15:04:05",
		"02-01-2006 15:04", "2-1-2006 15:04", "02-01-2006 15:04:05", "2-1-2006 15:04:05",
	}

	// Заменяем "/" на "." — часто из локалей
	sNorm := strings.ReplaceAll(s, "/", ".")

	for _, l := range layouts {
		if t, err := time.ParseInLocation(l, sNorm, time.UTC); err == nil {
			// обнулим время
			return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
		}
	}

	// Последний шанс — чистый ISO
	if t, err := time.ParseInLocation("2006-01-02", s, time.UTC); err == nil {
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
	}

	return time.Time{}, errors.New("unsupported date format")
}

func stripDecor(s string) string {
	s = strings.TrimSpace(s)
	// отрезаем по первому пробелу/букве T, если дальше идёт время/зона
	if idx := strings.IndexAny(s, "T "); idx != -1 {
		// только если после идёт что-то похожее на время
		tail := s[idx+1:]
		if strings.ContainsAny(tail, ":Z+-") {
			s = s[:idx]
		}
	}
	// выкидываем «г.»/«г» в конце
	s = strings.TrimSpace(strings.TrimSuffix(strings.TrimSuffix(s, " г."), " г"))
	s = strings.TrimSpace(strings.TrimSuffix(s, "г."))
	s = strings.TrimSpace(strings.TrimSuffix(s, "г"))
	return s
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
