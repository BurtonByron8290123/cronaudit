package monitor

import (
	"strings"
	"testing"
	"time"
)

const testYear = 2024

func TestParseSyslog_MatchesJobName(t *testing.T) {
	input := strings.NewReader(
		"Dec 15 10:05:01 myhost CRON[1234]: (root) CMD (/usr/local/bin/backup.sh)\n" +
			"Dec 15 11:00:00 myhost CRON[5678]: (root) CMD (/usr/local/bin/cleanup.sh)\n",
	)

	entries, err := ParseSyslog(input, []string{"/usr/local/bin/backup.sh"}, testYear)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].JobName != "/usr/local/bin/backup.sh" {
		t.Errorf("unexpected job name: %s", entries[0].JobName)
	}
	if entries[0].Status != "started" {
		t.Errorf("expected status 'started', got %s", entries[0].Status)
	}
}

func TestParseSyslog_FailedStatus(t *testing.T) {
	input := strings.NewReader(
		"Dec 15 10:05:01 myhost CRON[1234]: (root) FAILED (/usr/local/bin/backup.sh)\n",
	)

	entries, err := ParseSyslog(input, []string{"/usr/local/bin/backup.sh"}, testYear)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Status != "failed" {
		t.Errorf("expected status 'failed', got %s", entries[0].Status)
	}
}

func TestParseSyslog_NoMatch(t *testing.T) {
	input := strings.NewReader(
		"Dec 15 10:05:01 myhost CRON[1234]: (root) CMD (/usr/local/bin/other.sh)\n",
	)

	entries, err := ParseSyslog(input, []string{"/usr/local/bin/backup.sh"}, testYear)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestParseSyslog_TimestampParsed(t *testing.T) {
	input := strings.NewReader(
		"Dec 15 10:05:01 myhost CRON[1234]: (root) CMD (/usr/local/bin/backup.sh)\n",
	)

	entries, err := ParseSyslog(input, []string{"/usr/local/bin/backup.sh"}, testYear)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least one entry")
	}

	ts := entries[0].Timestamp
	if ts.Month() != time.December {
		t.Errorf("expected December, got %s", ts.Month())
	}
	if ts.Day() != 15 {
		t.Errorf("expected day 15, got %d", ts.Day())
	}
	if ts.Hour() != 10 || ts.Minute() != 5 || ts.Second() != 1 {
		t.Errorf("unexpected time: %v", ts)
	}
}

func TestParseSyslog_SkipsMalformedLines(t *testing.T) {
	input := strings.NewReader(
		"this is not a cron line\n" +
			"Dec 15 10:05:01 myhost CRON[1234]: (root) CMD (/usr/local/bin/backup.sh)\n",
	)

	entries, err := ParseSyslog(input, []string{"/usr/local/bin/backup.sh"}, testYear)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
}
