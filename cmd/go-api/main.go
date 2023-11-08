package main

import (
	"go-api/internal/config"
	productList "go-api/internal/http-server/handlers/product/list"
	productRead "go-api/internal/http-server/handlers/product/read"
	productSave "go-api/internal/http-server/handlers/product/save"
	productGroupList "go-api/internal/http-server/handlers/productGroup/list"
	productGroupRead "go-api/internal/http-server/handlers/productGroup/read"
	productGroupSave "go-api/internal/http-server/handlers/productGroup/save"
	"go-api/internal/http-server/handlers/redirect"
	"go-api/internal/http-server/handlers/url/remove"
	"go-api/internal/http-server/handlers/url/save"
	"go-api/internal/lib/logger/handlers/slogpretty"
	"go-api/internal/lib/logger/sl"
	"go-api/internal/storage/sqlite"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	// config: cleanenv
	cfg := config.MustLoad()

	// logger: slog
	log := setupLogger(cfg.Env)

	log.Info("starting go-api", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	// storage: sqlite
	storageSqlite, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	_ = storageSqlite

	// router: chi, chi-render
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	// own logger
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Get("/{alias}", redirect.New(log, storageSqlite))

	router.Route("/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth("go-api", map[string]string{
			cfg.HTTPServer.User: cfg.HTTPServer.Password,
		}))

		r.Post("/", save.New(log, storageSqlite))
		r.Delete("/{alias}",
			remove.New(log, storageSqlite))
	})

	router.Route("/product", func(r chi.Router) {
		r.Use(middleware.BasicAuth("go-api", map[string]string{
			cfg.HTTPServer.User: cfg.HTTPServer.Password,
		}))

		r.Post("/save", productSave.New(log, storageSqlite))
		r.Get("/{product_id}", productRead.New(log, storageSqlite))
		r.Post("/list", productList.New(log, storageSqlite))
	})

	router.Route("/productGroup", func(r chi.Router) {
		r.Use(middleware.BasicAuth("go-api", map[string]string{
			cfg.HTTPServer.User: cfg.HTTPServer.Password,
		}))

		r.Post("/save", productGroupSave.New(log, storageSqlite))
		r.Get("/{product_group_id}", productGroupRead.New(log, storageSqlite))
		r.Post("/list", productGroupList.New(log, storageSqlite))
	})

	log.Info("starting server", slog.String("address", cfg.Address))

	// server:
	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}

	log.Error("server stopped")
}

func setupLogger(env string) *slog.Logger {

	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
