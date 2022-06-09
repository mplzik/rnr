package rnr

import (
	"fmt"
	"sync"

	"github.com/mplzik/rnr/golang/pkg/pb"
	"google.golang.org/protobuf/proto"
)

// Nested Task

type NestedTaskCallback func(*NestedTask, *[]Task)

type NestedTask struct {
	pbMutex  sync.Mutex
	pb       pb.Task
	children []Task
	oldState map[*Task]pb.TaskState
	opts     NestedTaskOptions
}

type NestedTaskOptions struct {
	CustomPoll  NestedTaskCallback // a callback called each time a Poll() on NestedTask is called.
	Parallelism int                // the number of tasks to run in parallel; defaults to 1.
	CompleteAll bool               // if `true`, the NestedTask will attempt to run all tasks before transitioning to either SUCCEEDED or FAILED state.
}

func NewNestedTask(name string, opts NestedTaskOptions) *NestedTask {
	ret := &NestedTask{}
	ret.pb.Name = name
	ret.oldState = make(map[*Task]pb.TaskState)
	ret.opts = opts

	// Sanitize opts
	if ret.opts.Parallelism < 1 {
		ret.opts.Parallelism = 1
	}

	return ret
}

func (nt *NestedTask) Add(task Task) error {
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

	if nt.opts.CustomPoll != nil {
		nt.opts.CustomPoll(nt, &nt.children)
	}

	running := 0
	pending := []Task{}

	// Perform scheduling
	for i := range nt.children {
		child := nt.children[i]
		state := taskSchedState(child.Proto(nil))

		// Count running tasks
		if state == RUNNING {
			running++
		}

		if state == PENDING {
			pending = append(pending, child)
		}
	}

	// Add more running tasks, if applicable. Don't try to stop tasks -- these have been likely invoked manually.
	for (running < nt.opts.Parallelism) && len(pending) > 0 {
		// TODO: we shouldn't be mutating the PB this way; let's add a mutating function to the task interface.
		pending[0].SetState(pb.TaskState_RUNNING)
		pending = pending[1:]
		running++
	}

	// Poll the running or changed tasks
	for _, child := range nt.children {
		pb := child.Proto(nil)
		state := taskSchedState(pb)

		// Poll a task iff it's running or it has its state changed recently
		if state == RUNNING || pb.State != nt.oldState[&child] {
			nt.oldState[&child] = pb.State
			child.Poll()
		}

	}

	successCount := 0
	failedCount := 0
	doneCount := 0
	for _, child := range nt.children {
		cpb := child.Proto(nil)
		if cpb.State == pb.TaskState_SUCCESS {
			successCount++
		} else if cpb.State == pb.TaskState_FAILED {
			failedCount++
		}

		if taskSchedState(cpb) == DONE {
			doneCount++
		}
	}

	nt.Proto(func(pb *pb.Task) { pb.Message = fmt.Sprintf("%d/%d", successCount, len(nt.children)) })

	// Handle termination
	if !nt.opts.CompleteAll && failedCount > 0 {
		// Fail everything on a first failed task.
		nt.SetState(pb.TaskState_FAILED)
		return
	}

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

func (nt *NestedTask) GetChild(name string) Task {

	for _, child := range nt.children {
		if child.Proto(nil).Name == name {
			return child
		}
	}

	return nil
}
