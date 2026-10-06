package main

import (
	"context"
	_ "embed"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/indico/flashsale/internal/api"
	"github.com/indico/flashsale/internal/config"
	"github.com/indico/flashsale/internal/repository"
	"github.com/indico/flashsale/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/001_init.up.sql
var initSchema string

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", "err", err)
		os.Exit(1)
	}

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := waitForDB(rootCtx, cfg.DatabaseURL, 30*time.Second); err != nil {
		logger.Error("postgres not ready", "err", err)
		os.Exit(1)
	}

	poolCfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		logger.Error("parse DATABASE_URL", "err", err)
		os.Exit(1)
	}
	poolCfg.MaxConns = 20
	poolCfg.MinConns = 2
	poolCfg.MaxConnLifetime = 30 * time.Minute
	poolCfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(rootCtx, poolCfg)
	if err != nil {
		logger.Error("connect pool", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	repo := repository.New(pool)
	if err := repo.RunMigrations(rootCtx, initSchema); err != nil {
		logger.Error("run migrations", "err", err)
		os.Exit(1)
	}
	logger.Info("migrations applied")

	svc := service.New(repo, cfg.ReservationTTL)
	router := api.NewRouter(svc)

	// Background expiry sweeper.
	sweeperCtx, cancelSweeper := context.WithCancel(rootCtx)
	defer cancelSweeper()
	go runExpirySweeper(sweeperCtx, repo, cfg.SweeperInterval, logger)
	logger.Info("expiry sweeper started", "interval", cfg.SweeperInterval.String())

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server crashed", "err", err)
			stop()
		}
	}()

	<-rootCtx.Done()
	logger.Info("shutdown signal received, draining")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "err", err)
	} else {
		logger.Info("http server closed cleanly")
	}
	cancelSweeper()
	pool.Close()
	logger.Info("shutdown complete")
}

func waitForDB(ctx context.Context, url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		pool, err := pgxpool.New(ctx, url)
		if err != nil {
			lastErr = err
			time.Sleep(500 * time.Millisecond)
			continue
		}
		pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		err = pool.Ping(pingCtx)
		cancel()
		pool.Close()
		if err == nil {
			return nil
		}
		lastErr = err
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
	if lastErr == nil {
		lastErr = errors.New("timeout")
	}
	return lastErr
}

func runExpirySweeper(ctx context.Context, repo *repository.Repository, interval time.Duration, logger *slog.Logger) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			n, err := repo.SweepExpired(ctx)
			if err != nil {
				if !errors.Is(err, context.Canceled) {
					logger.Warn("sweeper failed", "err", err)
				}
				continue
			}
			if n > 0 {
				logger.Info("expired reservations swept", "count", n)
			}
		}
	}
}
