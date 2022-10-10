package rnr

import (
	"context"
	"sync"

	"github.com/mplzik/rnr/golang/pkg/pb"
	proto "google.golang.org/protobuf/proto"
)

type AsyncFunc func(context.Context, *AsyncTask)

type AsyncTask struct {
	pbMutex   sync.Mutex
	pb        pb.Task
	bgTask    AsyncFunc
	parentCtx context.Context
	ctx       context.Context
	cancel    context.CancelFunc

	// If set to `true`, the task will keep on running even after switching to SUCCESS state, making it a sort-of background task.
	runsInSuccess bool
}

func NewAsyncTask(name string, ctx context.Context, runsInSuccess bool, bgTask AsyncFunc) *AsyncTask {
	ret := &AsyncTask{}
	ret.pb.Name = name
	ret.bgTask = bgTask
	ret.parentCtx = ctx
	ret.ctx = nil
	ret.cancel = nil
	ret.runsInSuccess = runsInSuccess

	return ret
}

// Task is a generic interface for pollable tasks
func (bt *AsyncTask) Poll() {
	state := bt.Proto(nil)

	if state.State == pb.TaskState_RUNNING || (bt.runsInSuccess && state.State == pb.TaskState_SUCCESS) {
		if bt.ctx == nil {
			bt.ctx, bt.cancel = context.WithCancel(bt.parentCtx)
			go bt.bgTask(bt.ctx, bt)
		}
	} else {
		if bt.ctx != nil {
			bt.cancel()
			bt.ctx = nil
			bt.cancel = nil
		}
	}
}

func (bt *AsyncTask) Proto(updater func(*pb.Task)) *pb.Task {
	bt.pbMutex.Lock()
	defer bt.pbMutex.Unlock()

	if updater != nil {
		updater(&bt.pb)
	}
	ret := proto.Clone(&bt.pb).(*pb.Task)

	return ret
}

func (bt *AsyncTask) SetState(state pb.TaskState) {
	bt.Proto(func(pb *pb.Task) { pb.State = state })
}

func (bt *AsyncTask) GetChild(name string) Task {
	return nil
}
