package rnr

import (
	"fmt"
	"testing"
)

func TestNestedTask_Add(t *testing.T) {
	nt := NewNestedTask("nested task test", 1)

	for _, tn := range []string{"foo", "bar"} {
		ct := newMockTask(tn)
		if err := nt.Add(ct); err != nil {
			t.Fatalf("unexpected error when adding task %s: %v", tn, err)
		}
	}

	tn := "foo"
	ct := newMockTask(tn)
	if err := nt.Add(ct); err == nil {
		t.Fatalf("expecting error when adding task with repeated name %s", tn)
	}
}

func BenchmarkNestedTask_Add(b *testing.B) {
	for _, n := range []int{1, 5, 10, 20, 30, 40, 50, 100, 1000, 10000} {
		name := fmt.Sprintf("NestedTask - %6d children", n)
		tasks := make([]*mockTask, 0, n)
		for i := 0; i < n; i++ {
			tasks = append(tasks, newMockTask(fmt.Sprintf("task: %6d", i)))
		}

		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				nt := NewNestedTask(name, 0)
				for _, t := range tasks {
					nt.Add(t)
				}
			}
		})
	}
}

func TestNestedTask_GetChild(t *testing.T) {
	nt := NewNestedTask("nested task test", 1)
	ct1 := newMockTask("child 1")
	ct2 := newMockTask("child 2")

	nt.Add(ct1)
	nt.Add(ct2)

	if ct := nt.GetChild("child 1"); ct != ct1 {
		t.Errorf("expecting GetChild to return %v, got %v", ct1, ct)
	}
	if ct := nt.GetChild("child 2"); ct != ct2 {
		t.Errorf("expecting GetChild to return %v, got %v", ct2, ct)
	}

	if ct := nt.GetChild("foobar"); ct != nil {
		t.Errorf("expecting GetChild to return nil, got %v", ct)
	}
}
