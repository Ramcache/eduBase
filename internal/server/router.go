package server

import (
	httpSwagger "github.com/swaggo/http-swagger"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"eduBase/internal/config"
	"eduBase/internal/handlers"
	"eduBase/internal/repository"
	"eduBase/internal/service"
)

func NewRouter(cfg *config.Config, db *pgxpool.Pool) *chi.Mux {
	r := chi.NewRouter()

	// === ВСЕ middleware ДО роутов ===
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	// убираем хвостовой слэш и дублирующиеся слэши
	r.Use(middleware.StripSlashes)

	// CORS
	origins := cfg.CORSAllowedOrigins
	corsOpts := cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}
	if len(origins) == 1 && origins[0] == "*" {
		corsOpts.AllowedOrigins = []string{"*"}
		corsOpts.AllowCredentials = false
	}
	r.Use(cors.Handler(corsOpts))
	// === конец блока middleware ===

	// DI wiring
	coreRepo := repository.NewStudentCoreRepo(db)
	docsRepo := repository.NewDocumentsRepo(db)
	medRepo := repository.NewMedicalRepo(db)
	conRepo := repository.NewConsentsRepo(db)
	cntRepo := repository.NewContactsRepo(db)

	svc := service.NewStudentService(coreRepo, docsRepo, medRepo, conRepo, cntRepo)

	coreH := handlers.NewStudentsCoreHandler(svc)
	docsH := handlers.NewDocumentsHandler(svc)
	medH := handlers.NewMedicalHandler(svc)
	conH := handlers.NewConsentsHandler(svc)
	cntH := handlers.NewContactsHandler(svc)
	expH := handlers.NewExportHandler(svc)
	impH := handlers.NewImportHandler(svc)

	// routes
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok"))
	})
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Route("/api", func(r chi.Router) {
		r.Route("/students", func(r chi.Router) {
			r.Post("/", coreH.Create)
			r.Get("/", coreH.List)    // фильтры + пагинация
			r.Get("/{id}", coreH.Get) // aggregate view
			r.Put("/{id}", coreH.Update)
			r.Patch("/{id}", coreH.Update)
			r.Post("/{id}/contacts", cntH.Add)
		})
		r.Delete("/students/contacts/{id}", cntH.Delete)

		r.Put("/students/docs", medSafeWrap(docsH.Upsert))    // upsert документов
		r.Put("/students/medical", medSafeWrap(medH.Upsert))  // upsert мед.данных
		r.Put("/students/consents", medSafeWrap(conH.Upsert)) // upsert согласий

		r.Get("/students/export.xlsx", expH.ExportStudents)
		r.Get("/students/import/template.xlsx", impH.Template)
		r.Post("/students/import.xlsx", impH.Import)
	})

	return r
}

// опционально: тонкий враппер для единообразной обработки
func medSafeWrap(h http.HandlerFunc) http.HandlerFunc { return h }
