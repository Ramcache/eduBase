package main

import (
	"context"
	"eduBase/internal/models"
	"fmt"
	"github.com/go-chi/cors"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"

	"eduBase/config"
	"eduBase/internal/handlers"
	"eduBase/internal/logger"
	"eduBase/internal/middleware"
	"eduBase/internal/repository"
	"eduBase/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/jackc/pgx/v5"

	_ "eduBase/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title eduBase API
// @version 1.0
// @description База школ с ролями ROO и School.
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	cfg := config.Load()
	logg := logger.New(cfg.AppEnv)

	conn, err := pgx.Connect(context.Background(), cfg.DBURL)
	if err != nil {
		log.Fatal("db connect failed:", err)
	}
	defer conn.Close(context.Background())

	jwtAuth := jwtauth.New("HS256", []byte(cfg.JWTSecret), nil)

	// === Repositories ===
	userRepo := repository.NewUserRepository(conn)
	schoolRepo := repository.NewSchoolRepository(conn)
	classRepo := repository.NewClassRepository(conn)
	staffRepo := repository.NewStaffRepository(conn)
	studentRepo := repository.NewStudentRepository(conn)
	statsRepo := repository.NewStatsRepository(conn)

	// === Services ===
	authSvc := services.NewAuthService(userRepo, jwtAuth)
	schoolSvc := services.NewSchoolService(schoolRepo)
	classSvc := services.NewClassService(classRepo, schoolRepo)
	staffSvc := services.NewStaffService(staffRepo)
	studentSvc := services.NewStudentService(studentRepo, classRepo, schoolRepo)
	statsSvc := services.NewStatsService(statsRepo, schoolRepo)

	// === Handlers ===
	authHandler := handlers.NewAuthHandler(authSvc)
	rooHandler := handlers.NewRooHandler(authSvc, schoolRepo)
	rooSchoolHandler := handlers.NewRooSchoolHandler(schoolSvc)
	schoolSelfHandler := handlers.NewSchoolSelfHandler(schoolSvc)

	classHandler := handlers.NewClassHandler(classSvc)
	staffHandler := handlers.NewStaffHandler(staffSvc)
	studentHandler := handlers.NewStudentHandler(studentSvc)
	statsHandler := handlers.NewStatsHandler(statsSvc)

	// создаём дефолтного админа (ROO)
	CreateDefaultAdmin(context.Background(), userRepo, logg)

	// === Router ===
	r := chi.NewRouter()

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// JWT verifier (разбирает токен и кладёт клеймы в контекст)
	r.Use(middleware.JWTVerifier(jwtAuth))

	// Public
	r.Get("/docs/*", httpSwagger.WrapHandler)
	r.Get("/health", handlers.HealthHandler)

	// Auth
	r.Group(func(r chi.Router) {
		authHandler.Routes(r)
	})

	// ================================
	// ROO-only
	// ================================
	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticator(jwtAuth))
		r.Use(middleware.RequireRole("roo"))

		rooHandler.Routes(r)
		rooSchoolHandler.Routes(r)
	})

	// ================================
	// SCHOOL-only: работа со своей школой
	// ================================
	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticator(jwtAuth))
		r.Use(middleware.RequireRole("school"))

		schoolSelfHandler.Routes(r) // /school/me
	})

	// ================================
	// ROO или SCHOOL — общие сущности
	// ================================
	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticator(jwtAuth))
		r.Use(middleware.RequireAnyRole("roo", "school"))

		classHandler.Routes(r)
		staffHandler.Routes(r)
		studentHandler.Routes(r)
		statsHandler.Routes(r)
	})

	logg.Infof("📘 Swagger: http://localhost:%s/docs/index.html", cfg.AppPort)
	logg.Infof("✅ Server started on port %s", cfg.AppPort)
	log.Fatal(http.ListenAndServe(":"+cfg.AppPort, r))
}

func CreateDefaultAdmin(ctx context.Context, userRepo *repository.UserRepository, logg *zap.SugaredLogger) {
	const defaultEmail = "admin"
	const defaultPassword = "admin"

	user, err := userRepo.GetByEmail(ctx, defaultEmail)
	if err == nil && user != nil {
		logg.Infof("default admin already exists: %s", defaultEmail)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
	if err != nil {
		logg.Errorf("failed to hash default admin password: %v", err)
		return
	}

	u := &models.User{
		Email:    defaultEmail,
		Password: string(hash),
		Role:     "roo",
	}

	if err := userRepo.Create(ctx, u); err != nil {
		logg.Errorf("failed to create default admin: %v", err)
		return
	}

	logg.Infof("✅ Default admin created: %s / %s", defaultEmail, defaultPassword)
	fmt.Println("Default admin:", defaultEmail, defaultPassword)
}
