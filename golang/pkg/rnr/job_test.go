package rnr

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mplzik/rnr/golang/pkg/pb"
)

func TestJob(t *testing.T) {
	var pollCount int

	mt := newMockTask("root", pb.TaskState_RUNNING, &pollCount)

	j := NewJob(mt)

	ctx := context.Background()

	var err error

	if err = j.Start(ctx, 5*time.Millisecond); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err = j.Start(ctx, time.Millisecond); !errors.Is(err, ErrJobAlreadyStarted) {
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

	if err := j.Stop(); !errors.Is(err, ErrJobNotRunning) {
		t.Fatalf("expecting rnr.ErrJobNotRunning, got %v", err)
	}
}
