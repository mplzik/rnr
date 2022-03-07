package rnr

import (
	"context"
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
}

func TestCallbackTask_GetChild(t *testing.T) {
	c := NewCallbackTask("callback task test", nil).GetChild("foo")

	if c != nil {
		t.Fatalf("expecting GetChild to return nil, got %#v", c)
	}
}
