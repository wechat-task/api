package executor

import (
	"context"
	"time"

	"github.com/wechat-task/api/internal/logger"
)

// Scheduler runs the executor on a fixed interval.
type Scheduler struct {
	executor  *Executor
	interval  time.Duration
	batchSize int
	stopCh    chan struct{}
	doneCh    chan struct{}
}

// NewScheduler creates a new Scheduler.
func NewScheduler(executor *Executor, interval time.Duration, batchSize int) *Scheduler {
	return &Scheduler{
		executor:  executor,
		interval:  interval,
		batchSize: batchSize,
		stopCh:    make(chan struct{}),
		doneCh:    make(chan struct{}),
	}
}

// Start runs the scheduler loop. Call in a goroutine.
func (s *Scheduler) Start() {
	defer close(s.doneCh)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	logger.Infof("Scheduler started (interval=%s, batchSize=%d)", s.interval, s.batchSize)

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), s.interval)
			s.executor.RunBatch(ctx, s.batchSize)
			cancel()
		case <-s.stopCh:
			logger.Info("Scheduler stopping")
			return
		}
	}
}

// Stop signals the scheduler to stop and waits for it to finish.
func (s *Scheduler) Stop() {
	close(s.stopCh)
	select {
	case <-s.doneCh:
		logger.Info("Scheduler stopped")
	case <-time.After(s.interval * 2):
		logger.Warn("Scheduler stop timed out")
	}
}
