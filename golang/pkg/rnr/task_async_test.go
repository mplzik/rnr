package rnr

import (
	"context"
	"testing"
	"time"

	"github.com/mplzik/rnr/golang/pkg/pb"
)

const tick = 10 * time.Millisecond

func TestAsyncTask_Lifecycle(t *testing.T) {
	running := false
	at := NewAsyncTask("Test async task", context.Background(), false, func(ctx context.Context, at *AsyncTask) {
		running = true
		select {
		case <-ctx.Done():
			running = false
		}
		at.Proto(func(t *pb.Task) {
			t.State = pb.TaskState_SUCCESS
		})
	})

	at.Poll()
	time.Sleep(tick)
	if running != false {
		t.Errorf("async task expected not running was found running")
	}

	at.Proto(func(t *pb.Task) {
		t.State = pb.TaskState_RUNNING
	})
	at.Poll()
	time.Sleep(tick)
	if running != true {
		t.Errorf("async task expected running was found not running")
	}

	at.Proto(func(t *pb.Task) {
		t.State = pb.TaskState_SUCCESS
	})
	at.Poll()
	time.Sleep(tick)
	if running != false {
		t.Errorf("async task expected not running anymore was found running")
	}
}

func TestAsyncTask_BackgroundLifecycle(t *testing.T) {
	running := false
	at := NewAsyncTask("Test async task", context.Background(), true, func(ctx context.Context, at *AsyncTask) {
		running = true
		select {
		case <-ctx.Done():
			running = false
		}
		at.Proto(func(t *pb.Task) {
			t.State = pb.TaskState_SUCCESS
		})
	})

	at.Poll()
	time.Sleep(tick)
	if running != false {
		t.Errorf("async task expected not running was found running")
	}

	at.Proto(func(t *pb.Task) {
		t.State = pb.TaskState_RUNNING
	})
	at.Poll()
	time.Sleep(tick)
	if running != true {
		t.Errorf("async task expected running was found not running")
	}

	at.Proto(func(t *pb.Task) {
		t.State = pb.TaskState_SUCCESS
	})
	at.Poll()
	time.Sleep(tick)
	if running != true {
		t.Errorf("async task expected to be running in SUCCESS was found not running")
	}

	at.Proto(func(t *pb.Task) {
		t.State = pb.TaskState_FAILED
	})
	at.Poll()
	time.Sleep(tick)
	if running != false {
		t.Errorf("async task expected not running anymore was found running")
	}
}

func TestAsyncTask_EarlyExit(t *testing.T) {
	at := NewAsyncTask("Test async task", context.Background(), false, func(ctx context.Context, at *AsyncTask) {
	})

	at.Proto(func(t *pb.Task) {
		t.State = pb.TaskState_RUNNING
	})
	at.Poll()
	time.Sleep(tick)

	at.Proto(func(t *pb.Task) {
		t.State = pb.TaskState_SUCCESS
	})
	at.Poll()
	time.Sleep(10 * time.Millisecond)
}
