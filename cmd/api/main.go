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

	userRepo := repository.NewUserRepository(conn)
	schoolRepo := repository.NewSchoolRepository(conn)

	authSvc := services.NewAuthService(userRepo, jwtAuth)
	schoolSvc := services.NewSchoolService(schoolRepo)

	authHandler := handlers.NewAuthHandler(authSvc)
	rooHandler := handlers.NewRooHandler(authSvc, schoolRepo)
	rooSchoolHandler := handlers.NewRooSchoolHandler(schoolSvc)

	r := chi.NewRouter()
	r.Use(middleware.JWTVerifier(jwtAuth))

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Group(func(r chi.Router) { // public
		authHandler.Routes(r)
	})

	r.Group(func(r chi.Router) { // ROO only
		r.Use(middleware.Authenticator(jwtAuth))
		r.Use(middleware.RequireRole("roo"))
		rooHandler.Routes(r)
		rooSchoolHandler.Routes(r)
	})

	logg.Infof("📘 Swagger: http://localhost:%s/docs/index.html", cfg.AppPort)
	logg.Infof("✅ Server started on port %s", cfg.AppPort)
	log.Fatal(http.ListenAndServe(":"+cfg.AppPort, r))
}
