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

// Mock task useful for testing
var _ Task = &mockTask{} // quick compiler check mockTask fulfills the interface

type mockTask struct {
	pbTask     *pb.Task
	finalState pb.TaskState
	pollCount  int
}

// Not useful at the moment
func (m *mockTask) Poll() {
	m.pollCount += 1
	if m.pbTask.State != pb.TaskState_RUNNING {
		return
	}
	m.SetState(m.finalState)
}
func (m *mockTask) SetState(state pb.TaskState) {
	m.Proto(func(pb *pb.Task) { pb.State = state })
}
func (m *mockTask) GetChild(string) Task { return nil }

func (m *mockTask) Proto(updater func(*pb.Task)) *pb.Task {
	if updater != nil {
		updater(m.pbTask)
	}

	return m.pbTask
}

func newMockTask(name string) *mockTask {
	return &mockTask{
		pbTask: &pb.Task{
			Name:  name,
			State: pb.TaskState_PENDING,
		},
		finalState: pb.TaskState_SUCCESS,
	}
}

func newMockFailingTask(name string) *mockTask {
	return &mockTask{
		pbTask: &pb.Task{
			Name:  name,
			State: pb.TaskState_PENDING,
		},
		finalState: pb.TaskState_FAILED,
	}
}

func compareTaskStates(t *testing.T, tasks []Task, states []pb.TaskState) {
	if len(tasks) != len(states) {
		t.Errorf("`tasks` and `states` should have the same length (%d != %d)", len(tasks), len(states))
	}

	for i, task := range tasks {
		p := task.Proto(nil)

		if state := p.State; state != states[i] {
			t.Errorf("Task `%s` expected to be in state %s, but was in %s instead.", p.Name, states[i], state)
		}
	}
}
