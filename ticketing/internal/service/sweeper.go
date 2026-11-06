package service

import (
	"context"
	"log/slog"
	"time"
)

// DefaultSweepInterval is how often expired holds are released.
const DefaultSweepInterval = 30 * time.Second

const sweepTimeout = 10 * time.Second

// HoldExpirer is the store operation the sweeper needs.
type HoldExpirer interface {
	ExpireHolds(ctx context.Context) (int, error)
}

// Sweeper periodically expires overdue holds. The store takes an advisory lock
// so several API instances can run a sweeper safely.
type Sweeper struct {
	expirer  HoldExpirer
	logger   *slog.Logger
	interval time.Duration
}

// NewSweeper builds a sweeper. A nil logger uses slog.Default().
func NewSweeper(expirer HoldExpirer, logger *slog.Logger) *Sweeper {
	if logger == nil {
		logger = slog.Default()
	}

	return &Sweeper{
		expirer:  expirer,
		logger:   logger,
		interval: DefaultSweepInterval,
	}
}

// Run sweeps immediately, then on every tick until ctx is cancelled.
func (s *Sweeper) Run(ctx context.Context) {
	s.sweep(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.sweep(ctx)
		}
	}
}

func (s *Sweeper) sweep(parent context.Context) {
	ctx, cancel := context.WithTimeout(parent, sweepTimeout)
	defer cancel()

	expired, err := s.expirer.ExpireHolds(ctx)
	if err != nil {
		s.logger.Warn("hold sweep failed", "error", err)
		return
	}
	if expired > 0 {
		s.logger.Info("expired holds released", "orders", expired)
	}
}
