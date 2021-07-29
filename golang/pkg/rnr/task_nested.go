package rnr

import (
	"fmt"

	"github.com/mplzik/rnr/golang/pkg/pb"
	"google.golang.org/protobuf/proto"
)

// Nested Task

type NestedTask struct {
	pb          pb.Task
	children    []TaskInterface
	parallelism int
}

func NewNestedTask(name string, parallelism int) *NestedTask {
	ret := &NestedTask{}
	ret.pb.Name = name
	ret.parallelism = parallelism

	return ret
}

func (nt *NestedTask) Add(task TaskInterface) error {
	newName := task.GetProto().GetName()

	for _, child := range nt.children {
		if child.GetProto().Name == newName {
			return fmt.Errorf("task named '%s' already exists", child.GetProto().Name)
		}
	}
	nt.children = append(nt.children, task)
	task.SetState(pb.TaskState_PENDING)

	return nil
}

func (nt *NestedTask) Poll() {
	if taskSchedState(&nt.pb) != RUNNING {
		return
	}

	running := 0
	pending := []TaskInterface{}

	// Poll running tasks
	for i := range nt.children {
		child := nt.children[i]
		state := taskSchedState(child.GetProto())
		if state == RUNNING {
			child.Poll()
			if taskSchedState(child.GetProto()) == RUNNING {
				running++
			}
		}

		if state == PENDING {
			pending = append(pending, child)
		}
	}

	// Add more running tasks, if applicable
	for (running < nt.parallelism) && len(pending) > 0 {
		// TODO: we shouldn't be mutating the PB this way; let's add a mutating function to the task interface.
		pending[0].SetState(pb.TaskState_RUNNING)
		pending = pending[1:]
		running++
	}

	successCount := 0
	failedCount := 0
	doneCount := 0
	for _, child := range nt.children {
		if child.GetProto().State == pb.TaskState_SUCCESS {
			successCount++
		} else if child.GetProto().State == pb.TaskState_FAILED {
			failedCount++
		}

		if taskSchedState(child.GetProto()) == DONE {
			doneCount++
		}
	}

	nt.pb.Message = fmt.Sprintf("%d/%d", successCount, len(nt.children))

	// Handle termination
	if doneCount == len(nt.children) {
		if successCount == len(nt.children) {
			nt.pb.State = pb.TaskState_SUCCESS
		} else {
			nt.pb.State = pb.TaskState_FAILED
		}
	}
}

func (nt *NestedTask) GetProto() *pb.Task {
	ret := proto.Clone(&nt.pb).(*pb.Task)

	for _, child := range nt.children {
		ret.Children = append(ret.Children, child.GetProto())
	}
	return ret
}

func (nt *NestedTask) SetState(state pb.TaskState) {
	nt.pb.State = state
}

func (nt *NestedTask) GetChild(name string) TaskInterface {

	for _, child := range nt.children {
		if child.GetProto().Name == name {
			return child
		}
	}

	return nil
}
