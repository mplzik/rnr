package rnr

import (
	"context"
	"fmt"
	"os/exec"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/mplzik/rnr/golang/pkg/pb"
)

// Shell Task
var _ Task = &ShellTask{}

type ShellTask struct {
	pbMutex  sync.Mutex
	pb       pb.Task
	children []Task
	cmdName  string
	cmdArgs  []string
	err      chan error

	cmd *exec.Cmd
}

func NewShellTask(name, cmd string, args ...string) *ShellTask {
	ret := &ShellTask{}

	ret.pb.Name = name
	ret.cmdName = cmd
	ret.cmdArgs = args

	ret.cmd = exec.Command(ret.cmdName, ret.cmdArgs...)
	ret.err = make(chan error)

	return ret
}

func (ct *ShellTask) Poll(ctx context.Context) {
	if ct.cmd.Process == nil {
		// Not yet started, let's launch it first
		if err := ct.cmd.Start(); err != nil {
			ct.pb.Message = fmt.Sprintf("failed to start: %v", err)
			ct.pb.State = pb.TaskState_FAILED
			return
		}

		go func() { ct.err <- ct.cmd.Wait() }()

		ct.pb.State = pb.TaskState_RUNNING
		ct.pb.Message = "Started"
		return
	}

	if ct.pb.State != pb.TaskState_RUNNING {
		return
	}

	select {
	default:
		// still running
	case <-ctx.Done():
		if ct.cmd.ProcessState != nil {
			// process was already killed/finished
			return
		}

		if err := ct.cmd.Process.Kill(); err != nil {
			ct.pb.Message = fmt.Sprintf("cannot kill process: %v", err)
			ct.pb.State = pb.TaskState_FAILED
		}

	case err := <-ct.err:
		ct.pb.Message = "Exited"
		// The process has finished
		if err != nil {
			ct.pb.State = pb.TaskState_FAILED
			ct.pb.Message = err.Error()
		} else {
			ct.pb.State = pb.TaskState_SUCCESS
		}
	}

	// TODO: if deemed safe, terminating the process while
}

func (ct *ShellTask) Proto(updater func(*pb.Task)) *pb.Task {
	ct.pbMutex.Lock()
	defer ct.pbMutex.Unlock()

	if updater != nil {
		updater(&ct.pb)
	}
	ret := proto.Clone(&ct.pb).(*pb.Task)

	return ret
}

func (nt *ShellTask) SetState(state pb.TaskState) {
	nt.Proto(func(pb *pb.Task) { pb.State = state })
}

func (nt *ShellTask) GetChild(name string) Task {
	return nil
}
