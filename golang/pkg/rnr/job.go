package rnr

import (
	"fmt"
	"sync"
	"time"

	"github.com/mplzik/rnr/golang/pkg/pb"
	proto "google.golang.org/protobuf/proto"
)

type Job struct {
	pbMutex sync.Mutex
	job     pb.Job
	root    Task
	stop    chan struct{}
}

func NewJob(root Task) *Job {
	ret := &Job{
		job: pb.Job{
			Version: 1,
			Uuid:    "1235abcdef",
			Root:    nil,
		},
		root: root,
		stop: make(chan struct{}),
	}

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for exit := false; !exit; {
			select {
			case _, ok := <-ret.stop:
				if !ok {
					exit = true
				}
			case <-ticker.C:
				ret.root.Poll()
			}
		}
	}()

	return ret
}

func (j *Job) Proto(updater func(*pb.Job)) *pb.Job {
	j.pbMutex.Lock()
	defer j.pbMutex.Unlock()

	if updater != nil {
		updater(&j.job)
	}

	ret := proto.Clone(&j.job).(*pb.Job)
	ret.Root = j.root.Proto(nil)

	return ret
}

func (j *Job) Poll() {
	j.root.Poll()
}

func (j *Job) TaskRequest(r *pb.TaskRequest) error {
	var task = j.root

	if task == nil {
		return fmt.Errorf("root task not configured")
	}

	for _, i := range r.Path {
		task = task.GetChild(i)

		if task == nil {
			return fmt.Errorf("task %v not found", r.Path)
		}
	}

	if r.State != pb.TaskState_UNKNOWN {
		task.SetState(r.State)
	}

	return nil
}

// Start is a shortcut for setting the root task to "running" state.
func (j *Job) Start() {
	j.root.SetState(pb.TaskState_RUNNING)
}
