package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"
)

type fakeExpirer struct {
	calls atomic.Int32
	err   error
	fired chan struct{}
}

func (f *fakeExpirer) ExpireHolds(context.Context) (int, error) {
	f.calls.Add(1)
	if f.fired != nil {
		select {
		case f.fired <- struct{}{}:
		default:
		}
	}
	if f.err != nil {
		return 0, f.err
	}
	return 1, nil
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestSweeperSweeps(t *testing.T) {
	expirer := &fakeExpirer{}
	sweeper := NewSweeper(expirer, discardLogger())

	sweeper.sweep(context.Background())
	sweeper.sweep(context.Background())

	if got := expirer.calls.Load(); got != 2 {
		t.Fatalf("sweeps = %d, want 2", got)
	}
}

func TestSweeperLogsErrors(t *testing.T) {
	expirer := &fakeExpirer{err: errors.New("database is down")}
	sweeper := NewSweeper(expirer, discardLogger())

	sweeper.sweep(context.Background())

	if got := expirer.calls.Load(); got != 1 {
		t.Fatalf("sweeps = %d, want 1", got)
	}
}

func TestSweeperRunStopsOnCancel(t *testing.T) {
	expirer := &fakeExpirer{fired: make(chan struct{}, 1)}
	sweeper := NewSweeper(expirer, discardLogger())
	sweeper.interval = time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		sweeper.Run(ctx)
		close(done)
	}()

	select {
	case <-expirer.fired:
	case <-time.After(time.Second):
		t.Fatal("sweeper did not sweep on start")
	}

	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("sweeper did not stop after cancel")
	}
}
