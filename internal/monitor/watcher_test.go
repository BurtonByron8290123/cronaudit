package monitor

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestWatcher_EmitsNewLines(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "syslog*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	w := NewWatcher(f.Name(), 10*time.Millisecond)
	lines := make(chan string, 10)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = w.Watch(ctx, lines)
	}()

	time.Sleep(20 * time.Millisecond)
	_, _ = f.WriteString("hello watcher\n")
	_, _ = f.WriteString("second line\n")

	var got []string
	deadline := time.After(500 * time.Millisecond)
	for len(got) < 2 {
		select {
		case l := <-lines:
			got = append(got, l)
		case <-deadline:
			t.Fatalf("timeout waiting for lines, got %d", len(got))
		}
	}

	if got[0] != "hello watcher" {
		t.Errorf("expected 'hello watcher', got %q", got[0])
	}
	if got[1] != "second line" {
		t.Errorf("expected 'second line', got %q", got[1])
	}
}

func TestWatcher_StopsOnContextCancel(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "syslog*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	w := NewWatcher(f.Name(), 10*time.Millisecond)
	lines := make(chan string, 10)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- w.Watch(ctx, lines)
	}()

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("watcher did not stop after context cancel")
	}
}

func TestWatcher_FileNotFound(t *testing.T) {
	w := NewWatcher("/nonexistent/path/syslog.log", 10*time.Millisecond)
	lines := make(chan string, 1)
	ctx := context.Background()
	err := w.Watch(ctx, lines)
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestIndexNewline(t *testing.T) {
	cases := []struct {
		input []byte
		want  int
	}{
		{[]byte("hello\nworld"), 5},
		{[]byte("noeol"), -1},
		{[]byte("\n"), 0},
	}
	for _, tc := range cases {
		got := indexNewline(tc.input)
		if got != tc.want {
			t.Errorf("indexNewline(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}
