package monitor

import (
	"context"
	"os/exec"
	"time"
)

// ExecResult holds the outcome of a local command execution.
type ExecResult struct {
	JobName  string
	Command  string
	ExitCode int
	Duration time.Duration
	Err      error
	RunAt    time.Time
}

// Executor runs local commands and returns their results.
type Executor struct {
	timeout time.Duration
}

// NewExecutor creates an Executor with the given per-job timeout.
func NewExecutor(timeout time.Duration) *Executor {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &Executor{timeout: timeout}
}

// Run executes cmd via the shell and returns an ExecResult.
func (e *Executor) Run(ctx context.Context, jobName, command string) ExecResult {
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	start := time.Now()
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	err := cmd.Run()
	duration := time.Since(start)

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	return ExecResult{
		JobName:  jobName,
		Command:  command,
		ExitCode: exitCode,
		Duration: duration,
		Err:      err,
		RunAt:    start,
	}
}
