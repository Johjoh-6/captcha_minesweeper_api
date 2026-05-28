package main

import (
	db "captcha_sweeper/internal/database"
	"captcha_sweeper/internal/routes"
	"captcha_sweeper/internal/servers"
	"captcha_sweeper/internal/version"
	"context"
	"expvar"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	err := run(logger)
	if err != nil {
		trace := string(debug.Stack())
		logger.Error(err.Error(), "trace", trace)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	showVersion := flag.Bool("version", false, "display version and exit")

	flag.Parse()

	if *showVersion {
		fmt.Printf("version: %s\n", version.Get())
		return nil
	}

	// load config
	cfg := servers.LoadConfig()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return err
	}
	logger.Info("Configuration validation successful")

	store, err := db.NewStore(cfg.DB.Dsn)
	if err != nil {
		return fmt.Errorf("connect db: %w", err)
	}
	defer store.Close()
	logger.Info("Database connection established")

	// Publish expvar metrics
	expvar.NewString("version").Set(version.Get())
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))
	expvar.Publish("database", expvar.Func(func() any {
		stats := store.Pool.Stat()
		return map[string]any{
			"acquire_count":              stats.AcquireCount(),
			"acquire_duration":           stats.AcquireDuration().String(),
			"acquired_conns":             stats.AcquiredConns(),
			"canceled_acquire_count":     stats.CanceledAcquireCount(),
			"constructing_conns":         stats.ConstructingConns(),
			"empty_acquire_count":        stats.EmptyAcquireCount(),
			"empty_acquire_wait_time":    stats.EmptyAcquireWaitTime().String(),
			"idle_conns":                 stats.IdleConns(),
			"max_conns":                  stats.MaxConns(),
			"total_conns":                stats.TotalConns(),
			"new_conns_count":            stats.NewConnsCount(),
			"max_lifetime_destroy_count": stats.MaxLifetimeDestroyCount(),
			"max_idle_destroy_count":     stats.MaxIdleDestroyCount(),
		}
	}))
	expvar.Publish("timestamp", expvar.Func(func() any {
		return time.Now().Unix()
	}))
	expvar.Publish("app_metrics", expvar.Func(func() any {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		m, err := store.Queries.GetAppMetrics(ctx)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{
			"captchas_total":        m.CaptchasTotal,
			"captchas_solved":       m.CaptchasSolved,
			"captchas_unsolved":     m.CaptchasUnsolved,
			"captchas_last_hour":    m.CaptchasLastHour,
			"sessions_total":        m.SessionsTotal,
			"sessions_bots":         m.SessionsBots,
			"sessions_non_bots":     m.SessionsNonBots,
		}
	}))

	// Initialize application
	app := servers.NewApplication(cfg, store, logger)

	// Initialize router
	router := routes.NewRouter(app)

	// Start server
	return servers.RunServer(cfg.HttpPort, router, logger, &app.WG)
}
