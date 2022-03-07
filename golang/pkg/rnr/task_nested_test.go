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
