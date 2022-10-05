package rnr

import (
	"context"
	"fmt"
	"testing"

	"github.com/mplzik/rnr/golang/pkg/pb"
)

func TestCallbackTask_Poll(t *testing.T) {
	ctx := context.Background()

	var callsCount int
	fn := func(ctx context.Context, ct *CallbackTask) (bool, error) {
		callsCount++
		if callsCount == 1 {
			return false, nil
		}

		return true, nil
	}

	ct := NewCallbackTask("callback test", fn)

	{ // shouldn't call callback because task isn't running
		ct.Poll(ctx)

		if callsCount != 0 {
			t.Fatalf("callback task shouldn't have been called because it's not running (%d calls)", callsCount)
		}
	}

	{ // should call the callback when task is running without changing task state
		ct.SetState(pb.TaskState_RUNNING)

		ct.Poll(ctx)

		if callsCount < 1 {
			t.Fatal("expecting callback function to be invoked")
		}
		if s := ct.Proto(nil).State; s != pb.TaskState_RUNNING {
			t.Fatalf("expecting task state %v, got %v", pb.TaskState_RUNNING, s)
		}
	}

	{ // should change state when callback is done
		ct.Poll(ctx)

		if callsCount != 2 {
			t.Errorf("should have invoked the callback function twice (%d calls)", callsCount)
		}
		if s := ct.Proto(nil).State; s != pb.TaskState_SUCCESS {
			t.Fatalf("expecting task state %v, got %v", pb.TaskState_SUCCESS, s)
		}
	}

	{ // shouldn't call the callback
		oldCount := callsCount
		ct.Poll(ctx)

		if callsCount != oldCount {
			t.Errorf("expecting callback to be invoked %d times, got %d invokations", oldCount, callsCount)
		}
	}

	t.Run("callback returns error", func(t *testing.T) {
		var done bool
		fn := func(ctx context.Context, ct *CallbackTask) (bool, error) {
			return done, fmt.Errorf("done %t", done)
		}
		ct := NewCallbackTask("failing callback test", fn)

		ct.SetState(pb.TaskState_RUNNING)
		ct.Poll(ctx)

		pbt := ct.Proto(nil)
		if s := pbt.State; s != pb.TaskState_RUNNING {
			t.Errorf("shouldn't update task state: %v", s)
		}
		if exp := "done false"; pbt.Message != exp {
			t.Errorf("expeciing message to be %q, got %q", exp, pbt.Message)
		}

		done = true

		ct.Poll(ctx)

		pbt = ct.Proto(nil)
		if s := pbt.State; s != pb.TaskState_FAILED {
			t.Errorf("task state should be %v, got %v", pb.TaskState_FAILED, s)
		}
		if exp := "done true"; pbt.Message != exp {
			t.Errorf("expectiing message to be %q, got %q", exp, pbt.Message)
		}
	})

	t.Run("context", func(t *testing.T) {
		type keyType int
		key := keyType(1024)
		val := 42
		ctx := context.WithValue(context.Background(), key, val)

		fn := func(ctx context.Context, ct *CallbackTask) (bool, error) {
			if got := ctx.Value(key); got != val {
				t.Fatalf("expecting context key %T(%#[1]v) with value %v, got %v", key, val, got)
			}
			return true, nil
		}

		ct := NewCallbackTask("foo", fn)
		ct.SetState(pb.TaskState_RUNNING)
		ct.Poll(ctx)
	})
}

func TestCallbackTask_GetChild(t *testing.T) {
	c := NewCallbackTask("callback task test", nil).GetChild("foo")

	if c != nil {
		t.Fatalf("expecting GetChild to return nil, got %#v", c)
	}
}
