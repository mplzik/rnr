package rnr

import (
	"context"
	"os/exec"

	"github.com/mplzik/rnr/golang/pkg/pb"
)

func NewShellTask(name, command string, args ...string) *Task {
	cmd := exec.Command(command, args...)
	err := make(chan error)

	return NewTask(name, false, func(ctx context.Context, task *Task) {
		if cmd.Process == nil {
			// Not yet started, let's launch it first
			go func() { err <- cmd.Run() }()
			task.Proto(func(taskpb *pb.Task) *pb.Task {
				taskpb.Message = "Started"
				return taskpb
			})
		}

		select {
		default:
			// still running
		case err := <-err:
			task.Proto(func(taskpb *pb.Task) *pb.Task {
				taskpb.Message = "Exited"
				// The process has finished
				if err != nil {
					taskpb.State = pb.TaskState_FAILED
					taskpb.Message = err.Error()
				} else {
					taskpb.State = pb.TaskState_SUCCESS
				}
				return taskpb
			})
		}

	})
}
