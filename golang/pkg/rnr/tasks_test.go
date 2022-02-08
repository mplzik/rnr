package rnr

import (
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
}
