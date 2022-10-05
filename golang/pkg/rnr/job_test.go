package rnr_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mplzik/rnr/golang/pkg/pb"
	"github.com/mplzik/rnr/golang/pkg/rnr"
)

var _ rnr.Task = &task{}

type task struct {
	pollFn func(context.Context)
}

func (t *task) Poll(ctx context.Context)      { t.pollFn(ctx) }
func (t *task) SetState(pb.TaskState)         {}
func (t *task) Proto(func(*pb.Task)) *pb.Task { return nil }
func (*task) GetChild(string) rnr.Task        { return nil }

func TestJob(t *testing.T) {
	var pollCount int

	task := &task{
		pollFn: func(context.Context) {
			pollCount++
			t.Logf("poll %02d", pollCount)
		},
	}

	j := rnr.NewJob(task)

	ctx := context.Background()

	var err error

	if err = j.Start(ctx, 5*time.Millisecond); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err = j.Start(ctx, time.Millisecond); !errors.Is(err, rnr.ErrJobAlreadyStarted) {
		t.Fatalf("expecting ErrJobAlreadyStarted, got %v", err)
	}

	var stopErr error

	go func() {
		for {
			if pollCount == 5 {
				stopErr = j.Stop()
				break
			}
		}
	}()

	<-j.Wait()

	if stopErr != nil {
		t.Fatalf("unexpected error: %v", stopErr)
	}

	if err := j.Stop(); !errors.Is(err, rnr.ErrJobNotRunning) {
		t.Fatalf("expecting rnr.ErrJobNotRunning, got %v", err)
	}
}
