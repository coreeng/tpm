package lab

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type Runner interface {
	Run(ctx context.Context, name string, args ...string) error
	Output(ctx context.Context, name string, args ...string) ([]byte, error)
}

type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, name string, args ...string) error {
	// #nosec G204 -- the lab runner intentionally invokes local tools with argv, never through a shell.
	output, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	if err != nil {
		return commandError(name, args, output, err)
	}
	return nil
}

func (ExecRunner) Output(ctx context.Context, name string, args ...string) ([]byte, error) {
	// #nosec G204 -- the lab runner intentionally invokes local tools with argv, never through a shell.
	output, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	if err != nil {
		return output, commandError(name, args, output, err)
	}
	return output, nil
}

func commandError(name string, args []string, output []byte, err error) error {
	message := strings.TrimSpace(string(output))
	if message == "" {
		return fmt.Errorf("%s %s failed: %w", name, strings.Join(args, " "), err)
	}
	return fmt.Errorf("%s %s failed: %w\n%s", name, strings.Join(args, " "), err, message)
}

type Command struct {
	Name string
	Args []string
}

type FakeRunnerResponse struct {
	Output []byte
	Err    error
}

type FakeRunner struct {
	Commands  []Command
	Responses []FakeRunnerResponse
}

func NewFakeRunner() *FakeRunner {
	return &FakeRunner{}
}

func (r *FakeRunner) QueueResponse(output []byte, err error) {
	r.Responses = append(r.Responses, FakeRunnerResponse{Output: append([]byte(nil), output...), Err: err})
}

func (r *FakeRunner) Run(_ context.Context, name string, args ...string) error {
	r.record(name, args...)
	return r.nextResponse().Err
}

func (r *FakeRunner) Output(_ context.Context, name string, args ...string) ([]byte, error) {
	r.record(name, args...)
	response := r.nextResponse()
	return append([]byte(nil), response.Output...), response.Err
}

func (r *FakeRunner) record(name string, args ...string) {
	r.Commands = append(r.Commands, Command{Name: name, Args: append([]string(nil), args...)})
}

func (r *FakeRunner) nextResponse() FakeRunnerResponse {
	if len(r.Responses) == 0 {
		return FakeRunnerResponse{}
	}
	response := r.Responses[0]
	r.Responses = r.Responses[1:]
	return response
}
