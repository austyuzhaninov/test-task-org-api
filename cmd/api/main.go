package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/austyuzhaninov/test-task-org-api/internal/handler/respond"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/austyuzhaninov/test-task-org-api/internal/config"
	"github.com/austyuzhaninov/test-task-org-api/internal/handler"
	"github.com/austyuzhaninov/test-task-org-api/internal/middleware"
	"github.com/austyuzhaninov/test-task-org-api/internal/repository"
	"github.com/austyuzhaninov/test-task-org-api/internal/service"
	"github.com/austyuzhaninov/test-task-org-api/migrations"
	"github.com/austyuzhaninov/test-task-org-api/pkg/logger"
)

// @title           Organization API
// @version         1.0
// @description     REST API для управления организационной структурой компании
// @host            localhost:8080
// @BasePath  		/api/v1
func main() {
	log := logger.New()
	cfg := config.Load()
	resp := respond.New(log)

	// ── БД ───────────────────────────────────────────────────────────────────
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

	// ── Миграции ─────────────────────────────────────────────────────────────
	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		log.Error("goose set dialect", slog.String("err", err.Error()))
		os.Exit(1)
	}
	if err := goose.Up(sqlDB, "."); err != nil {
		log.Error("goose up", slog.String("err", err.Error()))
		os.Exit(1)
	}
	log.Info("migrations applied")

	// ── GORM ─────────────────────────────────────────────────────────────────
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		log.Error("failed to init gorm", slog.String("err", err.Error()))
		os.Exit(1)
	}

	// ── Dependency injection ──────────────────────────────────────────────────
	deptRepo := repository.NewDepartmentRepo(db)
	empRepo := repository.NewEmployeeRepo(db)

	deptSvc := service.NewDepartmentService(deptRepo, empRepo)
	empSvc := service.NewEmployeeService(empRepo, deptRepo)

	deptHandler := handler.NewDepartmentHandler(deptSvc, resp)
	empHandler := handler.NewEmployeeHandler(empSvc, resp)

	logMw := middleware.NewLogger(log)
	router := handler.NewRouter(deptHandler, empHandler, logMw)

	// ── HTTP сервер ───────────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

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
