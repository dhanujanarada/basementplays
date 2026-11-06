package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata"

	"github.com/dhanujanarada/basementplays/ticketing/internal/config"
	"github.com/dhanujanarada/basementplays/ticketing/internal/db"
	"github.com/dhanujanarada/basementplays/ticketing/internal/httpapi"
	"github.com/dhanujanarada/basementplays/ticketing/internal/qr"
	"github.com/dhanujanarada/basementplays/ticketing/internal/service"
	"github.com/dhanujanarada/basementplays/ticketing/internal/store"
)

// version is overridden with -ldflags "-X main.version=...".
var version = "dev"

func main() {
	if err := run(os.Args[1:]); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	command := "serve"
	if len(args) > 0 {
		command = args[0]
	}
	if len(args) > 1 {
		return fmt.Errorf("unexpected arguments: %v", args[1:])
	}

	switch command {
	case "serve", "migrate", "version":
	default:
		return fmt.Errorf("unknown command %q (want serve, migrate, or version)", command)
	}

	if command == "version" {
		fmt.Println(version)
		return nil
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.SlogLevel()}))
	slog.SetDefault(logger)

	if command == "migrate" {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		if err := db.Migrate(ctx, cfg.DatabaseURL, logger); err != nil {
			return err
		}
		logger.Info("migrations applied")
		return nil
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if cfg.RunMigrationsOnStart {
		migrateCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		err := db.Migrate(migrateCtx, cfg.DatabaseURL, logger)
		cancel()
		if err != nil {
			return fmt.Errorf("startup migrations: %w", err)
		}
	}

	pool, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	st := store.New(pool)

	sweeper := service.NewSweeper(st, logger)
	go sweeper.Run(ctx)

	orders := service.NewOrders(
		st,
		cfg.CheckoutHoldMinutes,
		cfg.ManualHoldHours,
		cfg.MaxSeatsPerOrder,
	)
	signer := qr.NewSigner(cfg.QRSecret, cfg.PublicBaseURL)
	tickets := service.NewTickets(st, signer)

	handler := httpapi.New(httpapi.Options{
		Logger:        logger,
		Pinger:        pool,
		Version:       version,
		InternalToken: cfg.InternalToken,
		Plans:         service.NewPlans(st, cfg.MaxPlanSeats),
		Shows:         service.NewShows(st),
		Orders:        orders,
		Availability:  service.NewAvailability(st),
		Holds:         service.NewHolds(st, cfg.HoldMinutes, cfg.MaxSeatsPerOrder),
		Tickets:       tickets,
		Scanner:       service.NewScanner(st, signer),

		CORSAllowedOrigins: cfg.CORSAllowedOrigins,
	})

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("http server listening", "addr", server.Addr, "version", version)
		serverErr <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("http server: %w", err)
		}
		return nil
	case <-ctx.Done():
		logger.Info("shutdown requested")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}

	logger.Info("shutdown complete")
	return nil
}
