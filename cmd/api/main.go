package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/austyuzhaninov/test-task-org-api/internal/config"
	"github.com/austyuzhaninov/test-task-org-api/pkg/logger"
	"github.com/pressly/goose/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	_ "github.com/lib/pq"
)

//go:embed ../../migrations/*.sql
var migrations embed.FS

func main() {
	log := logger.New()

	cfg := config.Load()

	// --- БД: сначала database/sql для goose ---
	sqlDB, err := sql.Open("postgres", cfg.DB.DSN())
	if err != nil {
		log.Error("failed to open db", slog.String("err", err.Error()))
		os.Exit(1)
	}
	defer sqlDB.Close()

	if err := waitForDB(sqlDB, log); err != nil {
		log.Error("db not ready", slog.String("err", err.Error()))
		os.Exit(1)
	}

	// --- Миграции через goose ---
	goose.SetBaseFS(migrations)
	if err := goose.SetDialect("postgres"); err != nil {
		log.Error("goose set dialect", slog.String("err", err.Error()))
		os.Exit(1)
	}
	if err := goose.Up(sqlDB, "migrations"); err != nil {
		log.Error("goose up", slog.String("err", err.Error()))
		os.Exit(1)
	}
	log.Info("migrations applied")

	// --- GORM поверх того же пула ---
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		log.Error("failed to init gorm", slog.String("err", err.Error()))
		os.Exit(1)
	}

	// TODO: wire handlers (следующие коммиты)
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"ok"}`)
	})

	_ = db // будет использован в следующих коммитах

	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info("server started", slog.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", slog.String("err", err.Error()))
			os.Exit(1)
		}
	}()

	<-quit
	log.Info("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("shutdown error", slog.String("err", err.Error()))
	}
}

// waitForDB ждёт готовности PostgreSQL (до 30 секунд).
func waitForDB(db *sql.DB, log *slog.Logger) error {
	for i := range 10 {
		if err := db.Ping(); err == nil {
			return nil
		}
		log.Info("waiting for db...", slog.Int("attempt", i+1))
		time.Sleep(3 * time.Second)
	}
	return fmt.Errorf("database not reachable after retries")
}
