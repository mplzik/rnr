package rnr

import (
	"fmt"
	"testing"

	"github.com/mplzik/rnr/golang/pkg/pb"
)

func TestNestedTask_Add(t *testing.T) {
	nt := NewNestedTask("nested task test", NestedTaskOptions{})

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
				nt := NewNestedTask(name, NestedTaskOptions{})
				for _, t := range tasks {
					nt.Add(t)
				}
			}
		})
	}
}

func TestNestedTask_GetChild(t *testing.T) {
	nt := NewNestedTask("nested task test", NestedTaskOptions{Parallelism: 1})
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
	nt := NewNestedTask("nested task test", NestedTaskOptions{Parallelism: 1, CompleteAll: false})
	ct1 := newMockFailingTask("child 1")
	ct2 := newMockTask("child 2")

	tasks := []Task{ct1, ct2, nt}

	nt.Add(ct1)
	nt.Add(ct2)
	nt.SetState(pb.TaskState_RUNNING)

	nt.Poll()
	compareTaskStates(t, tasks, []pb.TaskState{pb.TaskState_FAILED, pb.TaskState_PENDING, pb.TaskState_FAILED})
}

func TestNestedTask_CompleteAllFail(t *testing.T) {
	nt := NewNestedTask("nested task test", NestedTaskOptions{Parallelism: 1, CompleteAll: true})
	ct1 := newMockFailingTask("child 1")
	ct2 := newMockTask("child 2")

	tasks := []Task{ct1, ct2, nt}

	nt.Add(ct1)
	nt.Add(ct2)
	nt.SetState(pb.TaskState_RUNNING)

	nt.Poll()
	compareTaskStates(t, tasks, []pb.TaskState{pb.TaskState_FAILED, pb.TaskState_PENDING, pb.TaskState_RUNNING})

	nt.Poll()
	compareTaskStates(t, tasks, []pb.TaskState{pb.TaskState_FAILED, pb.TaskState_SUCCESS, pb.TaskState_FAILED})
}

func TestNestedTask_CompleteAllSuccess(t *testing.T) {
	ct1 := newMockTask("child 1")
	ct2 := newMockTask("child 2")
	nt := NewNestedTask("nested task test", NestedTaskOptions{Parallelism: 1, CompleteAll: true})

	tasks := []Task{ct1, ct2, nt}

	nt.Add(ct1)
	nt.Add(ct2)
	nt.SetState(pb.TaskState_RUNNING)

	nt.Poll()
	compareTaskStates(t, tasks, []pb.TaskState{pb.TaskState_SUCCESS, pb.TaskState_PENDING, pb.TaskState_RUNNING})

	nt.Poll()
	compareTaskStates(t, tasks, []pb.TaskState{pb.TaskState_SUCCESS, pb.TaskState_SUCCESS, pb.TaskState_SUCCESS})
}

func TestNestedTask_CallbackInvoked(t *testing.T) {
	childrenAdded := 0
	nt := NewNestedTask("nested task test", NestedTaskOptions{
		Parallelism: 1,
		CompleteAll: true,
		CustomPoll: func(nt *NestedTask, children *[]Task) {
			childName := fmt.Sprintf("callback-added child %d", childrenAdded)
			nt.Add(newMockTask(childName))
			childrenAdded += 1
		},
	})

	nt.SetState(pb.TaskState_PENDING)
	nt.Poll()
	if childrenAdded != 0 {
		t.Errorf("callback shouldn't be invoked for non-RUNNING nested task!")
	}

	nt.SetState(pb.TaskState_RUNNING)
	nt.Poll()
	if childrenAdded != 1 {
		t.Errorf("nested task callback was not invoked for running task!")
	}
	if len(nt.children) != 1 {
		t.Errorf("expected 1 child, got %d", len(nt.children))
	}

	nt.SetState(pb.TaskState_RUNNING)
	nt.Poll()
	fmt.Println(nt.children[1].Proto(nil).Name)
	if len(nt.children) != 2 {
		t.Errorf("expected 2 children, got %d", len(nt.children))
	}
}
