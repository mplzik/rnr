package rnr

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mplzik/rnr/golang/pkg/pb"
	proto "google.golang.org/protobuf/proto"
)

type Job struct {
	pbMutex  sync.Mutex
	job      pb.Job
	root     Task
	stop     chan struct{}
	oldProto *pb.Task
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
				ret.Poll()
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

func (j *Job) Poll() {
	j.root.Poll()

	newProto := j.root.Proto(nil)
	// Calculate diff and post state changes
	diff := taskDiff([]string{newProto.GetName()}, j.oldProto, newProto)

	log.Printf("State changed: %s\n", strings.Join(diff, "\n"))
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

// Start is a shortcut for setting the root task to "running" state.
func (j *Job) Start() {
	j.root.SetState(pb.TaskState_RUNNING)
}
