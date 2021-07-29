package rnr

import (
	"os/exec"

	proto "github.com/golang/protobuf/proto"
	"github.com/mplzik/rnr/golang/pkg/pb"
)

// Shell Task

type ShellTask struct {
	pb       pb.Task
	children []TaskInterface
	cmdName  string
	cmdArgs  []string
	err      chan error

	cmd *exec.Cmd
}

func NewShellTask(name string, cmdArg ...string) *ShellTask {
	ret := &ShellTask{}

	ret.pb.Name = name
	ret.cmdName = cmdArg[0]
	ret.cmdArgs = cmdArg[1:]

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

func (ct *ShellTask) GetProto() *pb.Task {
	ret := proto.Clone(&ct.pb).(*pb.Task)

	return ret
}

func (nt *ShellTask) SetState(state pb.TaskState) {
	nt.pb.State = state
}

func (nt *ShellTask) GetChild(name string) TaskInterface {
	return nil
}
