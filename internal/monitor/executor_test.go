package monitor

import (
	"context"
	"testing"
	"time"
)

func TestExecutor_Run_Success(t *testing.T) {
	ex := NewExecutor(5 * time.Second)
	res := ex.Run(context.Background(), "echo-job", "echo hello")

	if res.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", res.ExitCode)
	}
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if res.JobName != "echo-job" {
		t.Errorf("expected job name echo-job, got %s", res.JobName)
	}
	if res.Duration <= 0 {
		t.Error("expected positive duration")
	}
}

func TestExecutor_Run_Failure(t *testing.T) {
	ex := NewExecutor(5 * time.Second)
	res := ex.Run(context.Background(), "fail-job", "exit 2")

	if res.ExitCode != 2 {
		t.Fatalf("expected exit code 2, got %d", res.ExitCode)
	}
	if res.Err == nil {
		t.Fatal("expected non-nil error")
	}
}

func TestExecutor_Run_Timeout(t *testing.T) {
	ex := NewExecutor(100 * time.Millisecond)
	res := ex.Run(context.Background(), "slow-job", "sleep 10")

	if res.ExitCode == 0 {
		t.Fatal("expected non-zero exit code on timeout")
	}
	if res.Err == nil {
		t.Fatal("expected error on timeout")
	}
}

func TestExecutor_DefaultTimeout(t *testing.T) {
	ex := NewExecutor(0)
	if ex.timeout != 30*time.Second {
		t.Errorf("expected default timeout 30s, got %v", ex.timeout)
	}
}

func TestExecutor_Run_RunAtSet(t *testing.T) {
	ex := NewExecutor(5 * time.Second)
	before := time.Now()
	res := ex.Run(context.Background(), "ts-job", "true")
	after := time.Now()

	if res.RunAt.Before(before) || res.RunAt.After(after) {
		t.Errorf("RunAt %v not between %v and %v", res.RunAt, before, after)
	}
}
