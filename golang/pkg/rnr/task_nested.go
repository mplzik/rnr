package rnr

import (
	"fmt"
	"sync"

	"github.com/mplzik/rnr/golang/pkg/pb"
	"google.golang.org/protobuf/proto"
)

// Nested Task

type NestedTask struct {
	pbMutex     sync.Mutex
	pb          pb.Task
	children    []TaskInterface
	oldState    map[*TaskInterface]pb.TaskState
	parallelism int
}

func NewNestedTask(name string, parallelism int) *NestedTask {
	ret := &NestedTask{}
	ret.pb.Name = name
	ret.parallelism = parallelism
	ret.oldState = make(map[*TaskInterface]pb.TaskState)
	return ret
}

func (nt *NestedTask) Add(task TaskInterface) error {
	newName := task.Proto(nil).GetName()

	for _, child := range nt.children {
		if child.Proto(nil).Name == newName {
			return fmt.Errorf("task named '%s' already exists", child.Proto(nil).Name)
		}
	}
	nt.children = append(nt.children, task)
	task.SetState(pb.TaskState_PENDING)
	nt.oldState[&task] = pb.TaskState_PENDING

	return nil
}

func (nt *NestedTask) Poll() {

	if taskSchedState(nt.Proto(nil)) != RUNNING {
		return
	}

	running := 0
	pending := []TaskInterface{}

	// Poll running tasks
	for i := range nt.children {
		child := nt.children[i]
		pb := child.Proto(nil)
		state := taskSchedState(child.Proto(nil))

		// Poll a task iff it's running or it has its state changed recently
		if state == RUNNING || pb.State != nt.oldState[&child] {
			nt.oldState[&child] = pb.State
			child.Poll()
			if taskSchedState(child.Proto(nil)) == RUNNING {
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
		if child.Proto(nil).State == pb.TaskState_SUCCESS {
			successCount++
		} else if child.Proto(nil).State == pb.TaskState_FAILED {
			failedCount++
		}

		if taskSchedState(child.Proto(nil)) == DONE {
			doneCount++
		}
	}

	nt.Proto(func(pb *pb.Task) { pb.Message = fmt.Sprintf("%d/%d", successCount, len(nt.children)) })

	// Handle termination
	if doneCount == len(nt.children) {
		if successCount == len(nt.children) {
			nt.SetState(pb.TaskState_SUCCESS)
		} else {
			nt.SetState(pb.TaskState_FAILED)
		}
	}
}

func (nt *NestedTask) Proto(updater func(*pb.Task)) *pb.Task {
	nt.pbMutex.Lock()
	defer nt.pbMutex.Unlock()

	if updater != nil {
		updater(&nt.pb)
	}

	ret := proto.Clone(&nt.pb).(*pb.Task)

	for _, child := range nt.children {
		ret.Children = append(ret.Children, child.Proto(nil))
	}
	return ret
}

func (nt *NestedTask) SetState(state pb.TaskState) {
	nt.Proto(func(pb *pb.Task) { pb.State = state })
}

func (nt *NestedTask) GetChild(name string) TaskInterface {

	for _, child := range nt.children {
		if child.Proto(nil).Name == name {
			return child
		}
	}

	return nil
}
