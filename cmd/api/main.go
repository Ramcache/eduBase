package main

import (
	"context"
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
)

func main() {
	cfg := config.Load()
	logg := logger.New(cfg.AppEnv)

	conn, err := pgx.Connect(context.Background(), cfg.DBURL)
	if err != nil {
		log.Fatal("db connect failed:", err)
	}
	defer conn.Close(context.Background())

	// JWT
	jwtAuth := jwtauth.New("HS256", []byte(cfg.JWTSecret), nil)

	// === Repositories ===
	userRepo := repository.NewUserRepository(conn)
	schoolRepo := repository.NewSchoolRepository(conn)

	// === Services ===
	authSvc := services.NewAuthService(userRepo, jwtAuth)
	schoolSvc := services.NewSchoolService(schoolRepo)

	// === Handlers ===
	authHandler := handlers.NewAuthHandler(authSvc)
	rooHandler := handlers.NewRooHandler(authSvc)
	rooSchoolHandler := handlers.NewRooSchoolHandler(schoolSvc)

	// === Router ===
	r := chi.NewRouter()

	// Проверка подписи токена для всех запросов
	r.Use(middleware.JWTVerifier(jwtAuth))

	// === Публичные маршруты ===
	r.Group(func(r chi.Router) {
		authHandler.Routes(r)
	})

	// === Только для ROO ===
	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticator(jwtAuth))
		r.Use(middleware.RequireRole("roo"))
		rooHandler.Routes(r)
		rooSchoolHandler.Routes(r)
	})

	logg.Infof("✅ Server started on port %s", cfg.AppPort)
	log.Fatal(http.ListenAndServe(":"+cfg.AppPort, r))
}
