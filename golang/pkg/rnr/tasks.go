package rnr

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/mplzik/rnr/golang/pkg/pb"
	proto "google.golang.org/protobuf/proto"
)

type TaskState int

const (
	UNKNOWN TaskState = iota
	PENDING
	RUNNING
	DONE
)

var ErrNoChildrenAllowed = errors.New("no children allowed for this kind of task")

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

// TaskCallback is installed by higher-level construct, such as callback task or nested task.
type TaskCallback func(context.Context, *Task)
type StateUpdateCallback func(*pb.Task) *pb.Task

// Task is a generic interface for pollable tasks
type Task struct {
	cb           TaskCallback
	pb           *pb.Task
	children     []*Task
	has_children bool
}

func NewTask(name string, children bool, cb TaskCallback) *Task {
	return &Task{
		cb: cb,
		pb: &pb.Task{
			Name:  name,
			State: pb.TaskState_PENDING,
		},
		children:     []*Task{},
		has_children: children,
	}
}

func (task *Task) Poll(ctx context.Context) {
	task.cb(ctx, task)
}

func (task *Task) Proto(updater StateUpdateCallback) *pb.Task {

	oldState, ok := proto.Clone(task.pb).(*pb.Task)
	if !ok {
		log.Fatalf("Failed to clone proto")
	}

	if updater != nil {
		task.pb = updater(oldState)
	}

	// Rebuild the children protobufs.
	// This is terribly inefficient, but probably the easiest thing to do.
	task.pb.Children = make([]*pb.Task, len(task.children))
	for i, c := range task.children {
		task.pb.Children[i] = c.Proto(nil)
	}

	return task.pb
}

// SetState is a shortcut for atomically setting a state in the proto
func (task *Task) SetState(state pb.TaskState) {
	task.Proto(func(pb *pb.Task) *pb.Task {
		pb.State = state
		return pb
	})
}

// GetChild returns a child with the specified name
func (task *Task) GetChild(name string) *Task {
	for _, c := range task.children {
		if c.pb.Name == name {
			return c
		}
	}

	return nil
}

func (nt *Task) Add(task *Task) error {
	if !nt.has_children {
		return ErrNoChildrenAllowed
	}

	newName := task.Proto(nil).Name

	for _, child := range nt.children {
		if child.Proto(nil).Name == newName {
			return fmt.Errorf("task named '%s' already exists", child.Proto(nil).Name)
		}
	}
	nt.children = append(nt.children, task)

	return nil
}
