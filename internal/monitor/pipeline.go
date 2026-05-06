package monitor

import (
	"context"
	"log"
	"time"

	"github.com/cronaudit/cronaudit/internal/config"
)

// AlertFunc is called when a job execution result requires an alert.
type AlertFunc func(result ExecResult)

// Pipeline ties together the Executor, History, and StateStore for a set of jobs.
type Pipeline struct {
	cfg      *config.Config
	exec     *Executor
	history  *History
	state    *StateStore
	onAlert  AlertFunc
}

// NewPipeline constructs a Pipeline from the provided config and alert callback.
func NewPipeline(cfg *config.Config, onAlert AlertFunc) *Pipeline {
	return &Pipeline{
		cfg:     cfg,
		exec:    NewExecutor(30 * time.Second),
		history: NewHistory(100),
		state:   NewStateStore(),
		onAlert: onAlert,
	}
}

// RunJob executes a single named job from the config and processes the result.
func (p *Pipeline) RunJob(ctx context.Context, jobName string) {
	var command string
	for _, j := range p.cfg.Jobs {
		if j.Name == jobName {
			command = j.Command
			break
		}
	}
	if command == "" {
		log.Printf("pipeline: job %q not found in config", jobName)
		return
	}

	res := p.exec.Run(ctx, jobName, command)

	status := StatusOK
	if res.ExitCode != 0 {
		status = StatusFailed
	}

	p.history.Record(Record{
		JobName:   jobName,
		Timestamp: res.RunAt,
		Status:    status,
	})

	changed := p.state.Set(jobName, status)
	if changed && status == StatusFailed {
		if p.onAlert != nil {
			p.onAlert(res)
		}
	}

	log.Printf("pipeline: job=%s status=%s duration=%v exit=%d",
		jobName, status, res.Duration.Round(time.Millisecond), res.ExitCode)
}
