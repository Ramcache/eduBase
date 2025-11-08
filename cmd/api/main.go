package main

import (
	"context"
	"eduBase/internal/models"
	"fmt"
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

	// === Services ===
	authSvc := services.NewAuthService(userRepo, jwtAuth)
	schoolSvc := services.NewSchoolService(schoolRepo)
	classSvc := services.NewClassService(classRepo)

	// === Handlers ===
	authHandler := handlers.NewAuthHandler(authSvc)
	rooHandler := handlers.NewRooHandler(authSvc, schoolRepo)
	rooSchoolHandler := handlers.NewRooSchoolHandler(schoolSvc)
	classHandler := handlers.NewClassHandler(classSvc)
	CreateDefaultAdmin(context.Background(), userRepo, logg)
	// === Router ===
	r := chi.NewRouter()
	r.Use(middleware.JWTVerifier(jwtAuth))

	// Public
	r.Get("/swagger/*", httpSwagger.WrapHandler)
	r.Get("/health", handlers.HealthHandler)
	r.Group(func(r chi.Router) {
		authHandler.Routes(r)
	})

	// ROO-only
	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticator(jwtAuth))
		r.Use(middleware.RequireRole("roo"))
		rooHandler.Routes(r)
		rooSchoolHandler.Routes(r)
	})

	// ROO or School (shared)
	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticator(jwtAuth))
		r.Use(middleware.RequireAnyRole("roo", "school"))
		classHandler.Routes(r)
	})

	logg.Infof("📘 Swagger: http://localhost:%s/docs/index.html", cfg.AppPort)
	logg.Infof("✅ Server started on port %s", cfg.AppPort)
	log.Fatal(http.ListenAndServe(":"+cfg.AppPort, r))
}

func CreateDefaultAdmin(ctx context.Context, userRepo *repository.UserRepository, log *zap.SugaredLogger) {
	adminEmail := "admin"
	adminPass := "admin"
	adminRole := "roo"

	user, err := userRepo.FindByEmail(ctx, adminEmail)
	if err == nil && user != nil {
		log.Infow("admin_exists", "email", adminEmail)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(adminPass), bcrypt.DefaultCost)
	if err != nil {
		log.Errorw("bcrypt_failed", "err", err)
		return
	}

	admin := &models.User{
		Email:    adminEmail,
		Password: string(hash),
		Role:     adminRole,
	}

	if err := userRepo.Create(ctx, admin); err != nil {
		log.Errorw("create_admin_failed", "err", err)
		return
	}

	log.Infow("admin_created", "email", adminEmail, "role", adminRole)
	fmt.Println("✅ Admin user created: email=admin password=admin role=roo")
}
