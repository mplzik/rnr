package rnr

import (
	"github.com/mplzik/rnr/golang/pkg/pb"
)

const (
	PENDING = 0
	RUNNING = 1
	DONE    = 2
)

func taskSchedState(pbt *pb.Task) int {
	switch pbt.State {
	case pb.TaskState_FAILED, pb.TaskState_SKIPPED, pb.TaskState_SUCCESS:
		return DONE

	case pb.TaskState_PENDING:
		return PENDING

	case pb.TaskState_ACTION_NEEDED, pb.TaskState_RUNNING:
		return RUNNING
	}

	return 0 // What should we do here?
}

// TaskInterface is a generic interface for pollable tasks
type TaskInterface interface {
	Poll()
	Proto(updater func(*pb.Task)) *pb.Task
	SetState(pb.TaskState)
	GetChild(name string) TaskInterface
}
