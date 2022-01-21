package rnr

import (
	"github.com/mplzik/rnr/golang/pkg/pb"
	"google.golang.org/protobuf/proto"
)

// CallbackTask

// CallbackTask implements simple task with synchronously called callback.
// It returns a boolean indicating whether to transition into a final state and an error in case an error has happened. These values are used to best-effort-update the task's protobuf. If (false, nil) is supplied, the task state will be left untouched
type SimpleCallbackTask struct {
	pb       pb.Task
	callback func(*SimpleCallbackTask) (bool, error)
}

// NewSimpleCallbackTask returns a new callback task.
func NewSimpleCallbackTask(name string, callback func(*SimpleCallbackTask) (bool, error)) *SimpleCallbackTask {
	ret := &SimpleCallbackTask{}

	ret.pb.Name = name
	ret.callback = callback

	return ret
}

// Poll synchronously calls the callback
func (ct *SimpleCallbackTask) Poll() {
	if taskSchedState(&ct.pb) != RUNNING {
		return
	}

	ret, err := ct.callback(ct)

	if err != nil {
		ct.pb.State = pb.TaskState_FAILED
		ct.pb.Message = err.Error()
	} else if ret {
		ct.pb.State = pb.TaskState_SUCCESS
	}
}

func (ct *SimpleCallbackTask) GetProto() *pb.Task {
	ret := proto.Clone(&ct.pb).(*pb.Task)
	return ret
}

func (ct *SimpleCallbackTask) SetState(state pb.TaskState) {
	ct.pb.State = state
}

func (ct *SimpleCallbackTask) GetChild(name string) TaskInterface {
	return nil
}
