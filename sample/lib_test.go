package main

import "testing"

func TestModifyTest(t *testing.T) {
	want := "Hello World"
	got := "Hi there"
	ModifyText(&got)
	if got != want {
		t.Errorf("Got text = %q, want %q", got, want)
	}
}
