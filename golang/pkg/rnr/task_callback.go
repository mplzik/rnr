package rnr

import (
	"github.com/mplzik/rnr/golang/pkg/pb"
	"google.golang.org/protobuf/proto"
)

// CallbackTask

// CallbackTask implements simple task with synchronously called callback.
type CallbackTask struct {
	pb       pb.Task
	callback func() (bool, error)
}

// NewCallbackTask returns a new callback task
func NewCallbackTask(name string, callback func() (bool, error)) *CallbackTask {
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

	ret, err := ct.callback()

	if err != nil {
		ct.pb.State = pb.TaskState_FAILED
		ct.pb.Message = err.Error()
	} else if ret == true {
		ct.pb.State = pb.TaskState_SUCCESS
	}
}

func (ct *CallbackTask) GetProto() *pb.Task {
	ret := proto.Clone(&ct.pb).(*pb.Task)
	return ret
}

func (ct *CallbackTask) SetState(state pb.TaskState) {
	ct.pb.State = state
}

func (ct *CallbackTask) GetChild(name string) TaskInterface {
	return nil
}
