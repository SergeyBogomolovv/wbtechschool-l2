package app

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type App struct {
	srv *http.Server
	log *slog.Logger
}

type Handler interface {
	Init(r chi.Router)
}

func New(log *slog.Logger, host, port string, handlers ...Handler) *App {
	router := chi.NewRouter()
	router.Use(middleware.Logger)

	for _, h := range handlers {
		h.Init(router)
	}

	srv := &http.Server{
		Handler: router,
		Addr:    net.JoinHostPort(host, port),
	}

	return &App{srv: srv, log: log}
}

func (a *App) Start() {
	go func() {
		a.log.Info("starting server", slog.String("addr", a.srv.Addr))
		if err := a.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.log.Error("failed to start server")
			panic("failed to start server")
		}
	}()
}

func (a *App) Stop() {
	const shutdownTimeout = time.Second * 5
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := a.srv.Shutdown(ctx); err != nil {
		panic("failed to shutdown server")
	}
	a.log.Info("server stopped")
}
