package rnr

import (
	"context"

	"github.com/mplzik/rnr/golang/pkg/pb"
)

type TaskState int

const (
	UNKNOWN TaskState = iota
	PENDING
	RUNNING
	DONE
)

func taskSchedState(pbt *pb.Task) TaskState {
	switch pbt.State {
	case pb.TaskState_FAILED, pb.TaskState_SKIPPED, pb.TaskState_SUCCESS:
		return DONE

	case pb.TaskState_PENDING:
		return PENDING

	case pb.TaskState_ACTION_NEEDED, pb.TaskState_RUNNING:
		return RUNNING
	}

	return UNKNOWN
}

// Task is a generic interface for pollable tasks
type Task interface {
	Poll(ctx context.Context)
	Proto(updater func(*pb.Task)) *pb.Task
	SetState(pb.TaskState)
	GetChild(name string) Task
}
