package main

import (
	"os"

	"github.com/Rusich90/GophKeeper/internal/server/app"
	"github.com/Rusich90/GophKeeper/internal/server/config"
	"github.com/Rusich90/GophKeeper/internal/server/logger"
)

func main() {
	cfg := config.InitConfig()

	log := logger.InitLogger(cfg.Environment)

	application, err := app.NewApp(cfg, log)
	if err != nil {
		log.Error("Failed to initialize application", "error", err)
		os.Exit(1)
	}

	if err := application.Run(); err != nil {
		log.Error("Application error", "error", err)
		os.Exit(1)
	}

	log.Info("Application stopped")
}
