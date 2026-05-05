package monitor

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParsedSchedule holds the parsed cron expression fields.
type ParsedSchedule struct {
	Minute  string
	Hour    string
	Day     string
	Month   string
	Weekday string
}

// ParseCronExpr parses a standard 5-field cron expression and returns
// the expected interval duration between runs. Only simple cases are
// supported (wildcards and fixed values). Returns an error for
// expressions that cannot be reduced to a single interval.
func ParseCronExpr(expr string) (time.Duration, error) {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return 0, fmt.Errorf("invalid cron expression %q: expected 5 fields, got %d", expr, len(fields))
	}

	p := ParsedSchedule{
		Minute:  fields[0],
		Hour:    fields[1],
		Day:     fields[2],
		Month:   fields[3],
		Weekday: fields[4],
	}

	return estimateInterval(p)
}

// estimateInterval converts a ParsedSchedule into an approximate
// repeat interval. Supports */n step syntax and wildcard (*) fields.
func estimateInterval(p ParsedSchedule) (time.Duration, error) {
	// If all fields are wildcards, interval is 1 minute.
	if p.Minute == "*" && p.Hour == "*" && p.Day == "*" && p.Month == "*" && p.Weekday == "*" {
		return time.Minute, nil
	}

	// Handle */n in minute field with all others as wildcards.
	if strings.HasPrefix(p.Minute, "*/") && p.Hour == "*" && p.Day == "*" && p.Month == "*" && p.Weekday == "*" {
		n, err := strconv.Atoi(strings.TrimPrefix(p.Minute, "*/"))
		if err != nil || n <= 0 {
			return 0, fmt.Errorf("invalid step in minute field: %q", p.Minute)
		}
		return time.Duration(n) * time.Minute, nil
	}

	// Handle */n in hour field with minute wildcard.
	if p.Minute == "*" && strings.HasPrefix(p.Hour, "*/") && p.Day == "*" && p.Month == "*" && p.Weekday == "*" {
		n, err := strconv.Atoi(strings.TrimPrefix(p.Hour, "*/"))
		if err != nil || n <= 0 {
			return 0, fmt.Errorf("invalid step in hour field: %q", p.Hour)
		}
		return time.Duration(n) * time.Hour, nil
	}

	// Fixed minute, wildcard hour → runs every hour.
	if p.Minute != "*" && !strings.Contains(p.Minute, "/") &&
		p.Hour == "*" && p.Day == "*" && p.Month == "*" && p.Weekday == "*" {
		return time.Hour, nil
	}

	// Fixed minute and fixed hour → runs once per day.
	if p.Minute != "*" && !strings.Contains(p.Minute, "/") &&
		p.Hour != "*" && !strings.Contains(p.Hour, "/") &&
		p.Day == "*" && p.Month == "*" && p.Weekday == "*" {
		return 24 * time.Hour, nil
	}

	return 0, fmt.Errorf("unsupported cron expression %q: cannot determine interval", fmt.Sprintf("%s %s %s %s %s",
		p.Minute, p.Hour, p.Day, p.Month, p.Weekday))
}
