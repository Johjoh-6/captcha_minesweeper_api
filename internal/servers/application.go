package servers

import (
	db "captcha_sweeper/internal/database"
	"log/slog"
	"sync"
)

type Application struct {
	Config *Config
	Store  *db.Store
	Logger *slog.Logger
	Errors *Errors
	WG     sync.WaitGroup
}

func NewApplication(
	config *Config,
	store *db.Store,
	logger *slog.Logger,
) *Application {
	errs := NewErrors(logger)
	return &Application{
		Config: config,
		Store:  store,
		Logger: logger,
		Errors: errs,
		WG:     sync.WaitGroup{},
	}
}
