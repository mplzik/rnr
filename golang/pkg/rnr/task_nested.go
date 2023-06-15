package rnr

import (
	"context"
	"fmt"

	"github.com/mplzik/rnr/golang/pkg/pb"
)

// Nested Task

type NestedTaskCallback func(*Task, []*Task)

type NestedTaskOptions struct {
	CustomPoll  NestedTaskCallback // a callback called each time a Poll() on NestedTask is called.
	Parallelism int                // the number of tasks to run in parallel; defaults to 1.
	CompleteAll bool               // if `true`, the NestedTask will attempt to run all tasks before transitioning to either SUCCEEDED or FAILED state.
}

func NewNestedTask(name string, opts NestedTaskOptions) *Task {

	// Sanitize opts
	if opts.Parallelism < 1 {
		opts.Parallelism = 1
	}

	return NewTask(name, true, func(ctx context.Context, task *Task) {
		// Begin insert

		if taskSchedState(task.Proto(nil)) != RUNNING {
			return
		}

		if opts.CustomPoll != nil {
			opts.CustomPoll(task, task.children)
		}

		running := 0
		pending := []*Task{}

		// Perform scheduling
		for i := range task.children {
			child := task.children[i]
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
		for (running < opts.Parallelism) && len(pending) > 0 {
			// TODO: we shouldn't be mutating the PB this way; let's add a mutating function to the task interface.
			pending[0].SetState(pb.TaskState_RUNNING)
			pending = pending[1:]
			running++
		}

		// Poll the child tasks
		for _, child := range task.children {
			child.Poll(ctx)
		}

		successCount := 0
		failedCount := 0
		doneCount := 0
		for _, child := range task.children {
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

		task.Proto(func(pb *pb.Task) *pb.Task {
			pb.Message = fmt.Sprintf("%d/%d", successCount, len(task.children))
			return pb
		})

		// Handle termination
		if !opts.CompleteAll && failedCount > 0 {
			// Fail everything on a first failed task.
			task.SetState(pb.TaskState_FAILED)
			return
		}

		if doneCount == len(task.children) {
			if successCount == len(task.children) {
				task.SetState(pb.TaskState_SUCCESS)
			} else {
				task.SetState(pb.TaskState_FAILED)
			}
		}
	})

	// end insert
}
