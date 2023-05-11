package rnr

import (
	"context"
	"fmt"
	"testing"

	"github.com/mplzik/rnr/golang/pkg/pb"
)

func TestCallbackTask_Poll(t *testing.T) {
	var callsCount int
	fn := func(ctx context.Context, task *pb.Task) *pb.Task {
		callsCount++
		if callsCount == 1 {
			task.State = pb.TaskState_RUNNING
			return task
		}

		task.State = pb.TaskState_SUCCESS
		return task
	}

	ct := NewCallbackTask("callback test", fn)

	{ // shouldn't call callback because task isn't running
		ct.Poll()

		if callsCount != 0 {
			t.Fatalf("callback task shouldn't have been called because it's not running (%d calls)", callsCount)
		}
	}

	{ // should call the callback when task is running without changing task state
		ct.SetState(pb.TaskState_RUNNING)

		ct.Poll()

		if callsCount < 1 {
			t.Fatal("expecting callback function to be invoked")
		}
		if s := ct.Proto(nil).State; s != pb.TaskState_RUNNING {
			t.Fatalf("expecting task state %v, got %v", pb.TaskState_RUNNING, s)
		}
	}

	{ // should change state when callback is done
		ct.Poll()

		if callsCount != 2 {
			t.Errorf("should have invoked the callback function twice (%d calls)", callsCount)
		}
		if s := ct.Proto(nil).State; s != pb.TaskState_SUCCESS {
			t.Fatalf("expecting task state %v, got %v", pb.TaskState_SUCCESS, s)
		}
	}

	{ // should call the callback
		oldCount := callsCount
		ct.Poll()

		if callsCount != oldCount+1 {
			t.Errorf("expecting callback to be invoked %d times, got %d invocations", oldCount+1, callsCount)
		}
	}

	t.Run("callback returns error", func(t *testing.T) {
		var done bool
		fn := func(ctx context.Context, task *pb.Task) *pb.Task {
			if done {
				task.State = pb.TaskState_FAILED
			} else {
				task.State = pb.TaskState_RUNNING
			}
			task.Message = fmt.Sprintf("done %t", done)
			return task
		}
		ct := NewCallbackTask("failing callback test", fn)

		ct.SetState(pb.TaskState_RUNNING)
		ct.Poll()

		pbt := ct.Proto(nil)
		if s := pbt.State; s != pb.TaskState_RUNNING {
			t.Errorf("shouldn't update task state: %v", s)
		}
		if exp := "done false"; pbt.Message != exp {
			t.Errorf("expeciing message to be %q, got %q", exp, pbt.Message)
		}

		done = true

		ct.Poll()

		pbt = ct.Proto(nil)
		if s := pbt.State; s != pb.TaskState_FAILED {
			t.Errorf("task state should be %v, got %v", pb.TaskState_FAILED, s)
		}
		if exp := "done true"; pbt.Message != exp {
			t.Errorf("expecting message to be %q, got %q", exp, pbt.Message)
		}
	})
}

func TestCallbackTask_GetChild(t *testing.T) {
	c := NewCallbackTask("callback task test", nil).GetChild("foo")

	if c != nil {
		t.Fatalf("expecting GetChild to return nil, got %#v", c)
	}
}
