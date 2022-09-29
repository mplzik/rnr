package rnr

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mplzik/rnr/golang/pkg/pb"
	proto "google.golang.org/protobuf/proto"
)

// Job polling interval
var pollInterval = 5 * time.Second

var (
	ErrJobNotRunning     = errors.New("job is not running")
	ErrJobAlreadyStarted = errors.New("job was already started")
)

type Job struct {
	pbMutex  sync.Mutex
	job      pb.Job
	root     Task
	oldProto *pb.Task
	err      error
	done     chan struct{}
}

func NewJob(root Task) *Job {
	return &Job{
		job: pb.Job{
			Version: 1,
			Uuid:    "1235abcdef",
			Root:    nil,
		},
		root: root,
	}
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

// taskDiff recursively walks the task protobufs and calculates any differences
func taskDiff(path []string, old *pb.Task, new *pb.Task) []string {
	var ret []string

	oldState := "(new)"
	newState := "(deleted)"
	oldMessage := ""
	newMessage := ""

	if old != nil {
		oldState = old.State.String()
		oldMessage = old.GetMessage()
	}
	if new != nil {
		newState = new.State.String()
		newMessage = new.GetMessage()
	}

	if oldState != newState || oldMessage != newMessage {
		ret = append(ret, fmt.Sprintf("[%s]: %s (%s) -> %s (%s)", strings.Join(path, "/"), oldState, oldMessage, newState, newMessage))
	}

	// Check children
	childrenMap := make(map[string]struct{})
	oldChildren := make(map[string]*pb.Task)
	if old != nil {
		for _, c := range old.Children {
			childrenMap[c.Name] = struct{}{}
			oldChildren[c.Name] = c
		}
	}

	newChildren := make(map[string]*pb.Task)
	if new != nil {
		for _, c := range new.Children {
			childrenMap[c.Name] = struct{}{}
			newChildren[c.Name] = c
		}
	}

	children := make([]string, 0, len(childrenMap))

	// `children` is now a list of unique children names
	for key, _ := range childrenMap {
		children = append(children, key)
	}

	sort.Strings(children)

	for _, child := range children {
		oldChild, _ := oldChildren[child]
		newChild, _ := newChildren[child]
		taskName := "(unknown)"
		if newChild != nil {
			taskName = newChild.Name
		} else if oldChild != nil {
			taskName = oldChild.Name
		} else {
			// This shouldn't happen, since `children` is constructed from old and new children
		}
		ret = append(ret, taskDiff(append(path, taskName), oldChild, newChild)...)
	}

	return ret
}

func (j *Job) Poll(ctx context.Context) {
	j.root.Poll()

	newProto := j.root.Proto(nil)
	// Calculate diff and post state changes
	diff := taskDiff([]string{newProto.GetName()}, j.oldProto, newProto)

	if len(diff) > 0 {
		log.Printf("State changed: %s\n", strings.Join(diff, "\n"))
	}

	j.oldProto = proto.Clone(newProto).(*pb.Task)
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

// Err returns whatever error might have happened after Start.
func (j *Job) Err() error { return j.err }

// Start is a shortcut for setting the root task to "running" state.
func (j *Job) Start(ctx context.Context) error {
	if j.done != nil {
		return ErrJobAlreadyStarted
	}

	j.root.SetState(pb.TaskState_RUNNING)

	j.done = make(chan struct{})

	go func() {
		ticker := time.NewTicker(pollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-j.done:
				// job was stopped by calling stopFn
				return

			case <-ctx.Done():
				// job stopp due to context being done (e.g. cancelled or timed out)
				j.err = ctx.Err()
				return

			case <-ticker.C:
				j.Poll(ctx)
			}
		}
	}()

	return nil
}

// Stop stops the running job
func (j *Job) Stop() error {
	if j.done == nil {
		return ErrJobNotRunning
	}
	close(j.done)
	j.done = nil
	return nil
}

// Wait waits for the Job to finish.
func (j *Job) Wait() <-chan struct{} { return j.done }
