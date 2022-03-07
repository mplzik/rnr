package rnr

import (
	"context"
	"fmt"
	"testing"

	"github.com/mplzik/rnr/golang/pkg/pb"
)

func TestCallbackTask_Poll(t *testing.T) {
	var callsCount int
	fn := func(ctx context.Context, ct *CallbackTask) (bool, error) {
		callsCount++
		if callsCount == 1 {
			return false, nil
		}

		return true, nil
	}

	ct := NewCallbackTask("callback test", fn)

	ct.Poll()

	if callsCount < 1 {
		t.Fatal("expecting callback function to be invoked")
	}

	ct.Poll()

	if callsCount > 1 {
		t.Error("should have invoked the callback function once")
	}

	// changine state should call Poll again. This is something we
	// *shouldn't* know about it.
	ct.SetState(pb.TaskState_UNKNOWN)

	ct.Poll()

	if callsCount != 2 {
		t.Errorf("expecting callback to be invoked twice, got %d invokations", callsCount)
	}

	t.Run("callback returns error", func(t *testing.T) {
		var done bool
		fn := func(ctx context.Context, ct *CallbackTask) (bool, error) {
			return done, fmt.Errorf("done %t", done)
		}
		ct := NewCallbackTask("failing callback test", fn)

		ct.Poll()

		pbt := ct.Proto(nil)
		if s := pbt.State; s != pb.TaskState_RUNNING {
			t.Errorf("shouldn't update task state: %v", s)
		}
		if exp := "done false"; pbt.Message != exp {
			t.Errorf("expectiing message to be %q, got %q", exp, pbt.Message)
		}

		ct.SetState(pb.TaskState_UNKNOWN)
		done = true

		ct.Poll()

		pbt = ct.Proto(nil)
		if s := pbt.State; s != pb.TaskState_FAILED {
			t.Errorf("task state should be %v, got %v", pb.TaskState_FAILED, s)
		}
		if exp := "done true"; pbt.Message != exp {
			t.Errorf("expectiing message to be %q, got %q", exp, pbt.Message)
		}
	})
}

func TestCallbackTask_GetChild(t *testing.T) {
	c := NewCallbackTask("callback task test", nil).GetChild("foo")

	if c != nil {
		t.Fatalf("expecting GetChild to return nil, got %#v", c)
	}
}
