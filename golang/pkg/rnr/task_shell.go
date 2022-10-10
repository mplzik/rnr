package rnr

import (
	"os/exec"
	"sync"

	"github.com/mplzik/rnr/golang/pkg/pb"
	"google.golang.org/protobuf/proto"
)

// Shell Task

type ShellTask struct {
	pbMutex sync.Mutex
	pb      pb.Task
	cmdName string
	cmdArgs []string
	err     chan error

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

func (ct *ShellTask) Poll() {
	if ct.cmd.Process == nil {
		// Not yet started, let's launch it first
		go func() { ct.err <- ct.cmd.Run() }()
		ct.pb.Message = "Started"
	}

	select {
	default:
		// still running
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
