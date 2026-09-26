package app

import (
	"net/http"

	"github.com/Hani-SCV/payment-callback-assignment/internal/config"
	"github.com/Hani-SCV/payment-callback-assignment/internal/database"
)

type App struct {
	server *http.Server
}

func New(cfg *config.Config) (*App, error) {
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	dependencies := NewDependencies(db)
	mux := NewRouter(dependencies)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	return &App{
		server: server,
	}, nil
}

func (a *App) Run(addr string) error {
	a.server.Addr = addr

	return a.server.ListenAndServe()
}
