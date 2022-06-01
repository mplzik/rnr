package rnr

import (
	"fmt"
	"testing"

	"github.com/mplzik/rnr/golang/pkg/pb"
)

func TestNestedTask_Add(t *testing.T) {
	nt := NewNestedTask("nested task test", &NestedTaskOptions{Parallelism: 1})

	for _, tn := range []string{"foo", "bar"} {
		ct := newMockTask(tn)
		if err := nt.Add(ct); err != nil {
			t.Fatalf("unexpected error when adding task %s: %v", tn, err)
		}
	}

	tn := "foo"
	ct := newMockTask(tn)
	if err := nt.Add(ct); err == nil {
		t.Fatalf("expecting error when adding task with repeated name %s", tn)
	}
}

func BenchmarkNestedTask_Add(b *testing.B) {
	for _, n := range []int{1, 5, 10, 20, 30, 40, 50, 100, 1000, 10000} {
		name := fmt.Sprintf("NestedTask - %6d children", n)
		tasks := make([]*mockTask, 0, n)
		for i := 0; i < n; i++ {
			tasks = append(tasks, newMockTask(fmt.Sprintf("task: %6d", i)))
		}

		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				nt := NewNestedTask(name, &NestedTaskOptions{})
				for _, t := range tasks {
					nt.Add(t)
				}
			}
		})
	}
}

func TestNestedTask_GetChild(t *testing.T) {
	nt := NewNestedTask("nested task test", &NestedTaskOptions{Parallelism: 1})
	ct1 := newMockTask("child 1")
	ct2 := newMockTask("child 2")

	nt.Add(ct1)
	nt.Add(ct2)

	if ct := nt.GetChild("child 1"); ct != ct1 {
		t.Errorf("expecting GetChild to return %v, got %v", ct1, ct)
	}
	if ct := nt.GetChild("child 2"); ct != ct2 {
		t.Errorf("expecting GetChild to return %v, got %v", ct2, ct)
	}

	if ct := nt.GetChild("foobar"); ct != nil {
		t.Errorf("expecting GetChild to return nil, got %v", ct)
	}
}

func TestNestedTask_FailFirst(t *testing.T) {
	nt := NewNestedTask("nested task test", &NestedTaskOptions{Parallelism: 1, CompleteAll: false})
	ct1 := newMockFailingTask("child 1")
	ct2 := newMockTask("child 2")

	nt.Add(ct1)
	nt.Add(ct2)
	nt.SetState(pb.TaskState_RUNNING)

	nt.Poll()

	if ct1.pbTask.State != pb.TaskState_FAILED {
		t.Errorf("expecting child 1 to be in failed state, got %v", ct1.pbTask.State)
	}
	if ct2.pbTask.State != pb.TaskState_PENDING {
		t.Errorf("expecting child 2 to be in pending state, got %v", ct2.pbTask.State)
	}
	if nt.pb.State != pb.TaskState_FAILED {
		t.Errorf("expecting child 2 to be in failed state, got %v", nt.pb.State)
	}
}

func TestNestedTask_CompleteAllFail(t *testing.T) {
	nt := NewNestedTask("nested task test", &NestedTaskOptions{Parallelism: 1, CompleteAll: true})
	ct1 := newMockFailingTask("child 1")
	ct2 := newMockTask("child 2")

	nt.Add(ct1)
	nt.Add(ct2)
	nt.SetState(pb.TaskState_RUNNING)

	nt.Poll()

	if ct1.pbTask.State != pb.TaskState_FAILED {
		t.Errorf("expecting child 1 to be in failed state, got %v", ct1.pbTask.State)
	}
	if ct2.pbTask.State != pb.TaskState_PENDING {
		t.Errorf("expecting child 2 to be in pending state, got %v", ct2.pbTask.State)
	}
	if nt.pb.State != pb.TaskState_RUNNING {
		t.Errorf("expecting child 2 to be in running state, got %v", nt.pb.State)
	}

	nt.Poll()

	if ct1.pbTask.State != pb.TaskState_FAILED {
		t.Errorf("expecting child 1 to be in failed state, got %v", ct1.pbTask.State)
	}
	if ct2.pbTask.State != pb.TaskState_SUCCESS {
		t.Errorf("expecting child 2 to be in succeeded state, got %v", ct2.pbTask.State)
	}
	if nt.pb.State != pb.TaskState_FAILED {
		t.Errorf("expecting child 2 to be in failed state, got %v", nt.pb.State)
	}
}

func TestNestedTask_CompleteAllSuccess(t *testing.T) {
	ct1 := newMockTask("child 1")
	ct2 := newMockTask("child 2")
	nt := NewNestedTask("nested task test", &NestedTaskOptions{Parallelism: 1, CompleteAll: true})

	tasks := []Task{ct1, ct2, nt}

	nt.Add(ct1)
	nt.Add(ct2)
	nt.SetState(pb.TaskState_RUNNING)

	nt.Poll()
	compareTaskStates(t, tasks, []pb.TaskState{pb.TaskState_SUCCESS, pb.TaskState_PENDING, pb.TaskState_RUNNING})

	nt.Poll()
	compareTaskStates(t, tasks, []pb.TaskState{pb.TaskState_SUCCESS, pb.TaskState_SUCCESS, pb.TaskState_SUCCESS})
}
