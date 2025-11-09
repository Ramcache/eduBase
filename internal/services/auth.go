package services

import (
	"context"
	"eduBase/internal/utils"
	"errors"
	"github.com/go-chi/jwtauth/v5"
	"time"

	"eduBase/internal/models"
	"eduBase/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo     *repository.UserRepository
	jwt      *jwtauth.JWTAuth
	tokenExp time.Duration
}

func NewAuthService(repo *repository.UserRepository, jwt *jwtauth.JWTAuth) *AuthService {
	return &AuthService{repo: repo, jwt: jwt, tokenExp: 24 * time.Hour}
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	u, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		// не раскрываем, существует ли пользователь
		time.Sleep(150 * time.Millisecond)
		return "", errors.New("invalid email or password")
	}

	// сравнение пароля
	passBytes := []byte(password)
	dbPassBytes := []byte(u.Password)

	switch u.Role {
	case "roo":
		// админские — хэшированные
		if err := bcrypt.CompareHashAndPassword(dbPassBytes, passBytes); err != nil {
			time.Sleep(150 * time.Millisecond)
			return "", errors.New("invalid email or password")
		}
	default:
		// школы — пока в чистом виде (на проде можно захэшировать)
		if u.Password != password {
			time.Sleep(150 * time.Millisecond)
			return "", errors.New("invalid email or password")
		}
	}

	// базовые клеймы
	claims := map[string]interface{}{
		"user_id": u.ID,
		"role":    u.Role,
		"exp":     time.Now().Add(s.tokenExp).Unix(),
	}

	// если школа — добавляем school_name
	if u.Role == "school" {
		schoolRepo := repository.NewSchoolRepository(s.repo.DB())
		if school, err := schoolRepo.GetByUserID(ctx, u.ID); err == nil {
			claims["school_name"] = school.Name
			claims["school_id"] = school.ID
		}
	}

	// формируем JWT
	_, tokenStr, err := s.jwt.Encode(claims)
	if err != nil {
		return "", errors.New("failed to generate token")
	}

	return tokenStr, nil
}

func (s *AuthService) RegisterSchool(ctx context.Context, email, name, director string, schoolRepo *repository.SchoolRepository) (string, error) {
	// 1. генерируем пароль
	password, err := utils.GeneratePassword(8)
	if err != nil {
		return "", errors.New("failed to generate password")
	}

	// 2. создаём пользователя с ролью school (plain password)
	u := &models.User{
		Email:    email,
		Password: password, // сохраняем без хеша
		Role:     "school",
	}
	if err := s.repo.Create(ctx, u); err != nil {
		return "", err
	}

	// 3. ищем id
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return "", err
	}

	// 4. создаём школу
	school := &models.School{
		Name:     name,
		Director: director,
	}
	if err := schoolRepo.Create(ctx, school, user.ID); err != nil {
		return "", err
	}

	return password, nil
}
