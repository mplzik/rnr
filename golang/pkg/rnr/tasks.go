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
	m := map[pb.TaskState]int{
		pb.TaskState_FAILED:        DONE,
		pb.TaskState_PENDING:       PENDING,
		pb.TaskState_RUNNING:       RUNNING,
		pb.TaskState_SKIPPED:       DONE,
		pb.TaskState_SUCCESS:       DONE,
		pb.TaskState_ACTION_NEEDED: RUNNING,
	}

	return m[pbt.State]
}

// TaskInterface is a generic interface for pollable tasks
type TaskInterface interface {
	Poll()
	Proto(updater func(*pb.Task)) *pb.Task
	SetState(pb.TaskState)
	GetChild(name string) TaskInterface
}
