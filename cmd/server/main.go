package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/diyantoro/ticker-dashboard/internal/config"
	"github.com/diyantoro/ticker-dashboard/internal/model"
	"github.com/diyantoro/ticker-dashboard/internal/provider/binance"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server berhenti dengan error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	configPath := flag.String("config", "config.yaml", "path ke file konfigurasi")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	cfg, err := config.Load(*configPath)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ticks := make(chan model.Tick, 256)
	client := &binance.Client{
		BaseURL: cfg.ProviderURL,
		Logger:  logger,
	}
	done := make(chan error, 1)
	go func() {
		done <- client.Run(ctx, cfg.Symbols, ticks)
	}()

	logger.Info("ticker server dimulai", "provider", client.Name(), "symbols", cfg.Symbols)
	for {
		select {
		case tick := <-ticks:
			logger.Info("tick",
				"symbol", tick.Symbol,
				"price", tick.Price,
				"volume", tick.Volume,
				"timestamp", tick.Timestamp,
				"source", tick.Source,
			)
		case err := <-done:
			return err
		case <-ctx.Done():
			return <-done
		}
	}
}
