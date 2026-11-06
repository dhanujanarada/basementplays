// Command seed creates development fixtures. It is idempotent.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/dhanujanarada/basementplays/ticketing/internal/config"
	"github.com/dhanujanarada/basementplays/ticketing/internal/db"
	"github.com/dhanujanarada/basementplays/ticketing/internal/seatplan"
	"github.com/dhanujanarada/basementplays/ticketing/internal/store"
)

const (
	sampleKey   = "sample-50"
	sampleName  = "Sample 50 (placeholder)"
	sampleSeats = 50
	samplePrice = 250000 // Rs. 2500.00
)

func main() {
	if err := run(); err != nil {
		slog.Error("seed failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	st := store.New(pool)

	if _, err := st.GetPlan(ctx, sampleKey); err == nil {
		fmt.Printf("PLACEHOLDER seat plan %q already exists; nothing to do.\n", sampleKey)
		return nil
	} else if !errors.Is(err, store.ErrNotFound) {
		return fmt.Errorf("check existing seed plan: %w", err)
	}

	layout, err := seatplan.GeneratePlan(seatplan.Params{
		Name: sampleName,
		Key:  sampleKey,
		Tiers: []seatplan.TierInput{
			{Name: "Standard", PriceCents: samplePrice, SeatCount: sampleSeats},
		},
		Columns:    seatplan.DefaultColumns,
		Numbering:  seatplan.NumberingSequential,
		StageLabel: "Stage",
		MaxSeats:   cfg.MaxPlanSeats,
	})
	if err != nil {
		return fmt.Errorf("generate sample plan: %w", err)
	}

	raw, err := json.Marshal(layout)
	if err != nil {
		return fmt.Errorf("encode sample layout: %w", err)
	}

	if _, err := st.CreatePlan(ctx, sampleKey, sampleName, raw); err != nil {
		if errors.Is(err, store.ErrPlanKeyExists) {
			fmt.Printf("PLACEHOLDER seat plan %q already exists; nothing to do.\n", sampleKey)
			return nil
		}
		return fmt.Errorf("create sample plan: %w", err)
	}

	fmt.Printf(
		"Seeded PLACEHOLDER seat plan %q: %d seats, 1 tier \"Standard\", Rs. %s. "+
			"Replace it with a real plan before selling tickets.\n",
		sampleKey,
		len(layout.Seats),
		formatCents(samplePrice),
	)

	return nil
}

func formatCents(cents int) string {
	return fmt.Sprintf("%d.%02d", cents/100, cents%100)
}
