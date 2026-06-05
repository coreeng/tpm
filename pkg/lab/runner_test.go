package lab

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestFakeRunnerRecordsCommands(t *testing.T) {
	runner := NewFakeRunner()

	err := runner.Run(context.Background(), "docker", "build", "-t", "image", ".")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	want := []Command{{Name: "docker", Args: []string{"build", "-t", "image", "."}}}
	if !reflect.DeepEqual(runner.Commands, want) {
		t.Fatalf("Commands = %#v, want %#v", runner.Commands, want)
	}
}

func TestFakeRunnerRecordsOutputCommandsAndReturnsData(t *testing.T) {
	runner := NewFakeRunner()
	runner.QueueResponse([]byte("image-id"), nil)

	out, err := runner.Output(context.Background(), "docker", "image", "inspect", "image")
	if err != nil {
		t.Fatalf("Output returned error: %v", err)
	}
	if string(out) != "image-id" {
		t.Fatalf("Output = %q, want image-id", string(out))
	}

	want := []Command{{Name: "docker", Args: []string{"image", "inspect", "image"}}}
	if !reflect.DeepEqual(runner.Commands, want) {
		t.Fatalf("Commands = %#v, want %#v", runner.Commands, want)
	}
}

func TestFakeRunnerPropagatesQueuedErrors(t *testing.T) {
	runner := NewFakeRunner()
	runErr := errors.New("run failed")
	outputErr := errors.New("output failed")
	runner.QueueResponse(nil, runErr)
	runner.QueueResponse([]byte("details"), outputErr)

	if err := runner.Run(context.Background(), "docker", "build"); !errors.Is(err, runErr) {
		t.Fatalf("Run error = %v, want %v", err, runErr)
	}
	out, err := runner.Output(context.Background(), "docker", "push")
	if !errors.Is(err, outputErr) {
		t.Fatalf("Output error = %v, want %v", err, outputErr)
	}
	if string(out) != "details" {
		t.Fatalf("Output = %q, want details", string(out))
	}
}

func TestFakeRunnerRecordsArgumentSnapshots(t *testing.T) {
	runner := NewFakeRunner()
	args := []string{"build", "-t", "image"}

	if err := runner.Run(context.Background(), "docker", args...); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	args[0] = "push"

	want := []Command{{Name: "docker", Args: []string{"build", "-t", "image"}}}
	if !reflect.DeepEqual(runner.Commands, want) {
		t.Fatalf("Commands = %#v, want %#v", runner.Commands, want)
	}
}

func TestFakeRunnerUsesQueuedResponsesInOrder(t *testing.T) {
	runner := NewFakeRunner()
	runner.QueueResponse([]byte("first"), nil)
	runner.QueueResponse([]byte("second"), nil)

	first, err := runner.Output(context.Background(), "docker", "inspect", "first")
	if err != nil {
		t.Fatalf("first Output returned error: %v", err)
	}
	second, err := runner.Output(context.Background(), "docker", "inspect", "second")
	if err != nil {
		t.Fatalf("second Output returned error: %v", err)
	}

	if string(first) != "first" || string(second) != "second" {
		t.Fatalf("outputs = %q, %q; want first, second", string(first), string(second))
	}
}

func TestExecRunnerRunIncludesCommandOutputOnFailure(t *testing.T) {
	err := ExecRunner{}.Run(context.Background(), "sh", "-c", "printf 'chart pull failed' >&2; exit 7")
	if err == nil {
		t.Fatal("Run returned nil error for failing command")
	}
	if !strings.Contains(err.Error(), "chart pull failed") {
		t.Fatalf("Run error = %q, want command output", err.Error())
	}
	if !strings.Contains(err.Error(), "sh -c") {
		t.Fatalf("Run error = %q, want command context", err.Error())
	}
}

func TestExecRunnerOutputIncludesCommandOutputOnFailure(t *testing.T) {
	_, err := ExecRunner{}.Output(context.Background(), "sh", "-c", "printf 'registry unauthorized' >&2; exit 9")
	if err == nil {
		t.Fatal("Output returned nil error for failing command")
	}
	if !strings.Contains(err.Error(), "registry unauthorized") {
		t.Fatalf("Output error = %q, want command output", err.Error())
	}
}
