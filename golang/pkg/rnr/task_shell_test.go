package rnr

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/mplzik/rnr/golang/pkg/pb"
)

func TestMain(m *testing.M) {
	var (
		runTests = flag.Bool("run-tests", true, "whether to run the tests")
		exitCode = flag.Int("exit-code", 0, "exit code when not running the tests")
		sleep    = flag.Duration("sleep", 0*time.Second, "how long to sleep to simulate activity")
	)

	flag.Parse()

	if *runTests {
		os.Exit(m.Run())
	}

	t := time.NewTicker(10 * time.Millisecond)
	a := time.After(*sleep)

	var done bool

	for !done {
		select {
		case <-a:
			fmt.Println("DONE AFTER", *sleep)
			done = true
		case <-t.C:
			fmt.Println("TICK")
		}
	}

	os.Exit(*exitCode)
}

func TestShellTask(t *testing.T) {
	run := func(ctx context.Context, exitCode int, expectedState pb.TaskState, sleep time.Duration) {
		t.Helper()

		task := NewShellTask("foo", os.Args[0], "-run-tests=false", "-exit-code", strconv.Itoa(exitCode), "-sleep", sleep.String())

		for {
			state := task.Proto(nil).State
			if state == pb.TaskState_SUCCESS || state == pb.TaskState_FAILED {
				break
			}
			task.Poll(ctx)
			time.Sleep(50 * time.Millisecond)
		}

		if state := task.Proto(nil).State; state != expectedState {
			t.Fatalf("expecting %v state, got %v", expectedState, state)
		}

		pb := task.Proto(nil)
		t.Logf("task in state %v with message %q", pb.State, pb.Message)
	}

	ctx := context.Background()

	run(ctx, 0, pb.TaskState_SUCCESS, 0)
	run(ctx, 1, pb.TaskState_FAILED, 0)

	ctx2, cancel := context.WithCancel(ctx)
	cancel() // manually cancel the context
	run(ctx2, 0, pb.TaskState_FAILED, 5*time.Second)

	ctx3, cancel := context.WithCancel(ctx)
	cancel() // manually cancel the context
	run(ctx3, 1, pb.TaskState_FAILED, 5*time.Second)
}

func TestShellTask_GetChild(t *testing.T) {
	c := NewShellTask("shell task test", "").GetChild("foo")

	if c != nil {
		t.Fatalf("expecting GetChild to return nil, got %#v", c)
	}
}
