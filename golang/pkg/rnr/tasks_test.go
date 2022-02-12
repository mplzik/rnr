package rnr

import (
	"fmt"
	"testing"

	"github.com/mplzik/rnr/golang/pkg/pb"
)

func TestTaskSchedState(t *testing.T) {
	run := func(name string, exp TaskState, states ...pb.TaskState) {
		t.Run(name, func(t *testing.T) {
			for _, s := range states {
				got := taskSchedState(&pb.Task{State: s})
				if exp != got {
					t.Errorf("expecting task %v to map to %v, got %v", s, exp, got)
				}
			}
		})
	}

	run("DONE", DONE, pb.TaskState_FAILED, pb.TaskState_SKIPPED, pb.TaskState_SUCCESS)

	run("PENDING", PENDING, pb.TaskState_PENDING)

	run("RUNNING", RUNNING, pb.TaskState_ACTION_NEEDED, pb.TaskState_RUNNING)

	// test we defined results for all pb.TaskState
	for name, state := range pb.TaskState_value {
		t.Run(fmt.Sprintf("pb.TaskState_%s", name), func(t *testing.T) {
			got := taskSchedState(&pb.Task{State: pb.TaskState(state)})

			if name == "UNKNOWN" {
				if got != 0 {
					t.Errorf("UNKNOWN should be map to 0, got %v", got)
				}
			} else if got == 0 {
				t.Errorf("missing mapping for TaskState_%s", name)
			}
		})
	}
}
