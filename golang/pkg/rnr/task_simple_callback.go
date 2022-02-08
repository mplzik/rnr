package rnr

import (
	"context"
	"sync"
	"time"

	"github.com/mplzik/rnr/golang/pkg/pb"
	"google.golang.org/protobuf/proto"
)

// CallbackTask

// CallbackTask implements a task with synchronously called callback.
// It returns a boolean indicating whether to transition into a final state and an error in case an error has happened. These values are used to best-effort-update the task's protobuf. If (false, nil) is supplied, the task state will be left untouched
type CallbackTask struct {
	pbMutex  sync.Mutex
	pb       pb.Task
	callback func(*CallbackTask, context.Context) (bool, error)
}

// NewCallbackTask returns a new callback task.
func NewCallbackTask(name string, callback func(*CallbackTask, context.Context) (bool, error)) *CallbackTask {
	ret := &CallbackTask{}

	ret.pb.Name = name
	ret.callback = callback

	return ret
}

// Poll synchronously calls the callback
func (ct *CallbackTask) Poll() {
	if taskSchedState(&ct.pb) != RUNNING {
		return
	}

	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	ret, err := ct.callback(ct, ctx)

	if ret {
		if err == nil {
			ct.pb.State = pb.TaskState_SUCCESS
		} else {
			ct.pb.State = pb.TaskState_FAILED
			ct.pb.Message = err.Error()
		}
	} else {
		if err != nil {
			ct.pb.Message = err.Error()
		}
	}
}

func (ct *CallbackTask) Proto(updater func(*pb.Task)) *pb.Task {
	ct.pbMutex.Lock()
	defer ct.pbMutex.Unlock()

	if updater != nil {
		updater(&ct.pb)
	}
	ret := proto.Clone(&ct.pb).(*pb.Task)

	return ret
}

func (ct *CallbackTask) SetState(state pb.TaskState) {
	ct.Proto(func(pb *pb.Task) { pb.State = state })

	// An additional call to let the task know about state change
	go ct.Poll()
}

func (ct *CallbackTask) GetChild(name string) Task {
	return nil
}
