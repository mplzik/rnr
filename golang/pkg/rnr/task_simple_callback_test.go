package rnr

import (
	"context"
	"testing"
)

func TestCallbackTask_Poll(t *testing.T) {
	var callbackCalled bool
	fn := func(ctx context.Context, ct *CallbackTask) (bool, error) {
		callbackCalled = true
		return true, nil
	}

	ct := NewCallbackTask("callback test", fn)

	ct.Poll()

	if !callbackCalled {
		t.Fatal("expecting callback function to be invoked")
	}
}
