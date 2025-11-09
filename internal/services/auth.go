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
		return "", errors.New("invalid email or password")
	}

	switch u.Role {
	case "roo":
		if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
			return "", errors.New("invalid email or password")
		}

	default:
		if u.Password != password {
			return "", errors.New("invalid email or password")
		}
	}

	// ✅ Генерируем JWT
	_, tokenStr, _ := s.jwt.Encode(map[string]interface{}{
		"user_id": u.ID,
		"role":    u.Role,
		"exp":     time.Now().Add(s.tokenExp).Unix(),
	})
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
