package validators

import (
	"errors"
	"regexp"
	"time"
)

var (
	reClass   = regexp.MustCompile(`^\d{1,2}[А-ЯA-Z]$`)
	reSNILS1  = regexp.MustCompile(`^\d{3}-\d{3}-\d{3}\s\d{2}$`)
	reSNILS2  = regexp.MustCompile(`^\d{11}$`)
	rePassSer = regexp.MustCompile(`^\d{4}$`)
	rePassNum = regexp.MustCompile(`^\d{6}$`)
)

func BasicCore(last, first, birthDate, gender string, schoolID int, classLabel string, admissionYear int, regAddr, factAddr, status string) error {
	if last == "" || first == "" {
		return errors.New("Фамилия и Имя обязательны")
	}
	if birthDate == "" {
		return errors.New("Дата рождения обязательна")
	}
	if gender != "m" && gender != "f" {
		return errors.New("Пол должен быть m или f")
	}
	if schoolID <= 0 {
		return errors.New("school_id обязателен")
	}
	if !reClass.MatchString(classLabel) {
		return errors.New("Класс должен быть вида 7А/11Б")
	}
	thisYear := time.Now().Year()
	if admissionYear < 1990 || admissionYear > thisYear+1 {
		return errors.New("Некорректный год поступления")
	}
	if regAddr == "" || factAddr == "" {
		return errors.New("Адрес регистрации и проживания обязательны")
	}
	switch status {
	case "enrolled", "transferred", "graduated", "expelled":
	default:
		return errors.New("Некорректный статус")
	}
	return nil
}

// Conditional: SNILS always; passport if age >= 14y1m; else birth certificate.
func CheckDocs(birth time.Time, snils *string, passSer, passNum, birthCert *string) error {
	// SNILS required
	if snils == nil || *snils == "" || !(reSNILS1.MatchString(*snils) || reSNILS2.MatchString(*snils)) {
		return errors.New("СНИЛС обязателен и должен быть в формате 123-456-789 00 или 11 цифр")
	}

	ageThreshold := birth.AddDate(14, 1, 0)
	now := time.Now()
	if !now.Before(ageThreshold) {
		// need passport
		if passSer == nil || !rePassSer.MatchString(*passSer) || passNum == nil || !rePassNum.MatchString(*passNum) {
			return errors.New("Паспорт обязателен (серия 4 цифры, номер 6 цифр) с 14 лет и 1 месяца")
		}
	} else {
		// need birth certificate
		if birthCert == nil || *birthCert == "" {
			return errors.New("Свидетельство о рождении обязательно для младше 14 лет")
		}
	}
	return nil
}

func ParseYMD(ptr *string) (*time.Time, error) {
	if ptr == nil || *ptr == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", *ptr)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
