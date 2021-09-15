package rnr

import (
	"fmt"
	"time"

	proto "github.com/golang/protobuf/proto"
	"github.com/mplzik/rnr/golang/pkg/pb"
)

type Job struct {
	job    pb.Job
	root   TaskInterface
	active bool
}

func NewJob(root TaskInterface) *Job {
	ret := &Job{
		job: pb.Job{
			Version: 1,
			Uuid:    "1235abcdef",
			Root:    nil,
		},
		root:   root,
		active: false,
	}
	return ret
}

func (j *Job) GetProto() *pb.Job {
	ret := proto.Clone(&j.job).(*pb.Job)
	ret.Root = j.root.GetProto()

	return ret
}

func (j *Job) Poll() {
	j.root.Poll()
}

func (j *Job) TaskRequest(r *pb.TaskRequest) error {
	var task = j.root

	if task == nil {
		return fmt.Errorf("Root task not configured")
	}

	for _, i := range r.Path {
		task = task.GetChild(i)

		if task == nil {
			return fmt.Errorf("Task %v not found", r.Path)
		}
	}

	if r.State != pb.TaskState_UNKNOWN {
		task.SetState(r.State)
	}

	return nil
}

func (j *Job) Start() {
	j.root.SetState(pb.TaskState_RUNNING)
}

func (j *Job) Activate() {
	if j.active == true {
		return
	}
	j.active = true
	go func() {
		for range time.Tick(time.Second) {
			j.Poll()
		}
	}()
}
