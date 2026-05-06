package monitor

import (
	"context"
	"log"
	"os"
	"time"
)

// Watcher tails a syslog file and emits new lines on a channel.
type Watcher struct {
	filePath string
	pollInterval time.Duration
}

// NewWatcher creates a Watcher for the given syslog file path.
func NewWatcher(filePath string, pollInterval time.Duration) *Watcher {
	return &Watcher{
		filePath:     filePath,
		pollInterval: pollInterval,
	}
}

// Watch opens the file, seeks to the end, and emits new lines until ctx is cancelled.
func (w *Watcher) Watch(ctx context.Context, lines chan<- string) error {
	f, err := os.Open(w.filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Seek to end so we only process new entries.
	if _, err := f.Seek(0, 2); err != nil {
		return err
	}

	buf := make([]byte, 0, 4096)
	tmp := make([]byte, 4096)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		n, _ := f.Read(tmp)
		if n == 0 {
			time.Sleep(w.pollInterval)
			continue
		}

		buf = append(buf, tmp[:n]...)
		for {
			idx := indexNewline(buf)
			if idx < 0 {
				break
			}
			line := string(buf[:idx])
			buf = buf[idx+1:]
			if line != "" {
				select {
				case lines <- line:
				case <-ctx.Done():
					return nil
				}
			}
		}
	}
}

func indexNewline(b []byte) int {
	for i, c := range b {
		if c == '\n' {
			return i
		}
	}
	return -1
}

// WatchAndProcess tails the syslog file and feeds matching records into the pipeline.
func WatchAndProcess(ctx context.Context, w *Watcher, jobNames []string, pipeline *Pipeline) {
	lines := make(chan string, 64)
	go func() {
		if err := w.Watch(ctx, lines); err != nil {
			log.Printf("watcher error: %v", err)
		}
		close(lines)
	}()

	for line := range lines {
		records := ParseSyslog([]string{line}, jobNames)
		for _, r := range records {
			pipeline.RunJob(r.JobName, r)
		}
	}
}
