package monitor

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/yourorg/cronaudit/internal/alert"
	"github.com/yourorg/cronaudit/internal/config"
)

func TestWatchAndProcess_Integration(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "syslog*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	// Pre-populate so the file exists; watcher will seek to end.
	_, _ = f.WriteString("# existing content\n")

	alerted := make(chan string, 4)
	notifier := &captureNotifier{ch: alerted}

	cfg := &config.Config{
		Jobs: []config.Job{
			{Name: "backup", Schedule: "0 2 * * *", Command: ""},
		},
	}

	h := NewHistory(5)
	st := NewStateStore()
	exec := NewExecutor(0)
	pipeline := NewPipeline(cfg, h, st, exec, notifier)

	w := NewWatcher(f.Name(), 10*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go WatchAndProcess(ctx, w, []string{"backup"}, pipeline)
	time.Sleep(30 * time.Millisecond)

	// Write a failed syslog entry for the backup job.
	ts := time.Now().Format("Jan  2 15:04:05")
	line := fmt.Sprintf("%s myhost CRON[1234]: (root) CMD (backup) FAILED\n", ts)
	_, _ = f.WriteString(line)

	select {
	case name := <-alerted:
		if name != "backup" {
			t.Errorf("expected alert for 'backup', got %q", name)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout: no alert received from WatchAndProcess")
	}
}

// captureNotifier records job names for which alerts were sent.
type captureNotifier struct {
	ch chan string
}

func (c *captureNotifier) Send(jobName, message string) error {
	c.ch <- jobName
	return nil
}
