package main

import (
	"log"
	"net/http"

	"eduBase/config"
	"eduBase/internal/logger"
	"eduBase/internal/server"
)

func main() {
	cfg := config.Load()
	logg := logger.New(cfg.AppEnv)

	r := server.NewRouter()
	logg.Infof("Server running on port %s", cfg.AppPort)
	log.Fatal(http.ListenAndServe(":"+cfg.AppPort, r))
}
