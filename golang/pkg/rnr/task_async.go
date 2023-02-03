package rnr

import (
	"context"

	"github.com/mplzik/rnr/golang/pkg/pb"
)

// type AsyncFunc func(context.Context, *AsyncTask)

// type AsyncTask struct {
// 	pbMutex   sync.Mutex
// 	pb        pb.Task
// 	bgTask    AsyncFunc
// 	parentCtx context.Context
// 	ctx       context.Context
// 	cancel    context.CancelFunc

// 	// If set to `true`, the task will keep on running even after switching to SUCCESS state, making it a sort-of background task.
// 	runsInSuccess bool
// }

// func NewAsyncTask(name string, ctx context.Context, runsInSuccess bool, bgTask AsyncFunc) *AsyncTask {
// 	ret := &AsyncTask{}
// 	ret.pb.Name = name
// 	ret.bgTask = bgTask
// 	ret.parentCtx = ctx
// 	ret.ctx = nil
// 	ret.cancel = nil
// 	ret.runsInSuccess = runsInSuccess

// 	return ret
// }

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
