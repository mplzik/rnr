package rnr

import (
	"context"
	"testing"
	"time"

	"github.com/mplzik/rnr/golang/pkg/pb"
)

const tick = 10 * time.Millisecond

func TestAsyncTask_Lifecycle(t *testing.T) {
	ctx := context.TODO()
	running := false
	at := NewAsyncTask("Test async task", context.Background(), false, func(ctx context.Context, update func(StateUpdateCallback) *pb.Task) {
		running = true
		select {
		case <-ctx.Done():
			running = false
		}
		update(func(t *pb.Task) *pb.Task {
			t.State = pb.TaskState_SUCCESS
			return t
		})
	})

	at.Poll(ctx)
	time.Sleep(tick)
	if running != false {
		t.Errorf("async task expected not running was found running")
	}

	at.Proto(func(t *pb.Task) *pb.Task {
		t.State = pb.TaskState_RUNNING
		return t
	})
	at.Poll(ctx)
	time.Sleep(tick)
	if running != true {
		t.Errorf("async task expected running was found not running")
	}

	at.Proto(func(t *pb.Task) *pb.Task {
		t.State = pb.TaskState_SUCCESS
		return t
	})
	at.Poll(ctx)
	time.Sleep(tick)
	if running != false {
		t.Errorf("async task expected not running anymore was found running")
	}
}

func TestAsyncTask_BackgroundLifecycle(t *testing.T) {
	ctx := context.TODO()
	running := false
	at := NewAsyncTask("Test async task", context.Background(), true, func(ctx context.Context, update func(StateUpdateCallback) *pb.Task) {
		running = true
		select {
		case <-ctx.Done():
			running = false
		}
		update(func(t *pb.Task) *pb.Task {
			t.State = pb.TaskState_SUCCESS
			return t
		})
	})

	at.Poll(ctx)
	time.Sleep(tick)
	if running != false {
		t.Errorf("async task expected not running was found running")
	}

	at.Proto(func(t *pb.Task) *pb.Task {
		t.State = pb.TaskState_RUNNING
		return t
	})
	at.Poll(ctx)
	time.Sleep(tick)
	if running != true {
		t.Errorf("async task expected running was found not running")
	}

	at.Proto(func(t *pb.Task) *pb.Task {
		t.State = pb.TaskState_SUCCESS
		return t
	})
	at.Poll(ctx)
	time.Sleep(tick)
	if running != true {
		t.Errorf("async task expected to be running in SUCCESS was found not running")
	}

	at.Proto(func(t *pb.Task) *pb.Task {
		t.State = pb.TaskState_FAILED
		return t
	})
	at.Poll(ctx)
	time.Sleep(tick)
	if running != false {
		t.Errorf("async task expected not running anymore was found running")
	}
}

func TestAsyncTask_EarlyExit(t *testing.T) {
	ctx := context.TODO()

	at := NewAsyncTask("Test async task", context.Background(), false, func(context.Context, func(StateUpdateCallback) *pb.Task) {
	})

	at.Proto(func(t *pb.Task) *pb.Task {
		t.State = pb.TaskState_RUNNING
		return t
	})
	at.Poll(ctx)
	time.Sleep(tick)

	at.Proto(func(t *pb.Task) *pb.Task {
		t.State = pb.TaskState_SUCCESS
		return t
	})
	at.Poll(ctx)
	time.Sleep(10 * time.Millisecond)
}
