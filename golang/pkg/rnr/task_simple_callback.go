package rnr

import (
	"context"

	"github.com/mplzik/rnr/golang/pkg/pb"
)

type CallbackFunc func(context.Context, *pb.Task) *pb.Task

// NewCallbackTask returns a new callback task.
func NewCallbackTask(name string, callback CallbackFunc) *Task {

	newTask := NewTask(name, false, func(ctx context.Context, task *Task) {
		var oldState = pb.TaskState_PENDING
		task.Proto(func(taskState *pb.Task) *pb.Task {
			if (taskState.State != pb.TaskState_RUNNING) && (oldState == taskState.State) {
				return taskState
			}

			oldState = taskState.State
			return callback(ctx, taskState)
		})
	})

	return newTask
}
