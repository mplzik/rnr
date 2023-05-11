package rnr

import (
	"context"

	"github.com/mplzik/rnr/golang/pkg/pb"
)

type AsyncFunc func(context.Context, func(StateUpdateCallback) *pb.Task)

func NewAsyncTask(name string, ctx context.Context, runsInSuccess bool, bgTask AsyncFunc) *Task {
	parentCtx := ctx
	var currentCtx context.Context
	var cancel context.CancelFunc

	ret := NewTask(name, false, func(ctx context.Context, task *Task) {
		state := task.Proto(nil)

		if state.State == pb.TaskState_RUNNING || (runsInSuccess && state.State == pb.TaskState_SUCCESS) {
			if currentCtx == nil {
				currentCtx, cancel = context.WithCancel(parentCtx)
				go bgTask(currentCtx, func(cb StateUpdateCallback) *pb.Task {
					return task.Proto(cb)
				})
			}
		} else {
			if currentCtx != nil {
				cancel()
				currentCtx = nil
				cancel = nil
			}
		}

	})

	return ret
}
