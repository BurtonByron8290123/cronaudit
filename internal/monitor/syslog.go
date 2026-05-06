package monitor

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
)

// SyslogEntry represents a parsed cron-related syslog line.
type SyslogEntry struct {
	JobName   string
	Timestamp time.Time
	Status    string // "started" | "finished" | "failed"
	Raw       string
}

// syslogPattern matches lines like:
// Dec 15 10:05:01 hostname CRON[12345]: (root) CMD (/usr/local/bin/backup.sh)
var syslogPattern = regexp.MustCompile(
	`^(\w+\s+\d+\s+\d{2}:\d{2}:\d{2})\s+\S+\s+CRON\[\d+\]:\s+\(\S+\)\s+(CMD|FAILED)\s+\((.+)\)$`,
)

// ParseSyslog reads syslog-formatted lines from r and returns cron entries
// whose command matches any of the provided job names.
func ParseSyslog(r io.Reader, jobNames []string, year int) ([]SyslogEntry, error) {
	nameSet := make(map[string]struct{}, len(jobNames))
	for _, n := range jobNames {
		nameSet[n] = struct{}{}
	}

	var entries []SyslogEntry
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		entry, ok := parseSyslogLine(line, year)
		if !ok {
			continue
		}
		if _, matched := nameSet[entry.JobName]; matched {
			entries = append(entries, entry)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("syslog scan error: %w", err)
	}
	return entries, nil
}

func parseSyslogLine(line string, year int) (SyslogEntry, bool) {
	matches := syslogPattern.FindStringSubmatch(strings.TrimSpace(line))
	if matches == nil {
		return SyslogEntry{}, false
	}

	timeStr := fmt.Sprintf("%s %d", matches[1], year)
	t, err := time.Parse("Jan  2 15:04:05 2006", timeStr)
	if err != nil {
		// Try single-digit day without extra space
		t, err = time.Parse("Jan _2 15:04:05 2006", timeStr)
		if err != nil {
			return SyslogEntry{}, false
		}
	}

	status := "started"
	if matches[2] == "FAILED" {
		status = "failed"
	}

	return SyslogEntry{
		JobName:   matches[3],
		Timestamp: t,
		Status:    status,
		Raw:       line,
	}, true
}
