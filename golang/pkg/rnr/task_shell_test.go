package rnr

import "testing"

func TestShellTask_GetChild(t *testing.T) {
	c := NewShellTask("shell task test", "").GetChild("foo")

	if c != nil {
		t.Fatalf("expecting GetChild to return nil, got %#v", c)
	}
}
