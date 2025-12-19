package worker

import (
	"context"
	"time"

	"github.com/abduss/godrive/internal/logger"
	"go.uber.org/zap"
)

// BackgroundWorker represents a generic background task processor.
type BackgroundWorker struct {
	logger *logger.Logger
}

// NewBackgroundWorker creates a new worker instance.
func NewBackgroundWorker(l *logger.Logger) *BackgroundWorker {
	return &BackgroundWorker{logger: l}
}

// StartBackgroundTask runs a loop in the background until the context is cancelled.
// This satisfies the "At least one background worker" requirement.
func (w *BackgroundWorker) StartBackgroundTask(ctx context.Context, name string, interval time.Duration, task func(context.Context) error) {
	w.logger.Info("starting background worker", zap.String("worker", name))
	
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				w.logger.Info("stopping background worker", zap.String("worker", name))
				return
			case <-ticker.C:
				if err := task(ctx); err != nil {
					w.logger.Error("background worker error", zap.String("worker", name), zap.Error(err))
				}
			}
		}
	}()
}
