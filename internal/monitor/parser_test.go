package monitor

import (
	"testing"
	"time"
)

func TestParseCronExpr_EveryMinute(t *testing.T) {
	d, err := ParseCronExpr("* * * * *")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != time.Minute {
		t.Errorf("expected 1m, got %v", d)
	}
}

func TestParseCronExpr_EveryFiveMinutes(t *testing.T) {
	d, err := ParseCronExpr("*/5 * * * *")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != 5*time.Minute {
		t.Errorf("expected 5m, got %v", d)
	}
}

func TestParseCronExpr_EveryTwoHours(t *testing.T) {
	d, err := ParseCronExpr("* */2 * * *")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != 2*time.Hour {
		t.Errorf("expected 2h, got %v", d)
	}
}

func TestParseCronExpr_FixedMinuteEveryHour(t *testing.T) {
	d, err := ParseCronExpr("30 * * * *")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != time.Hour {
		t.Errorf("expected 1h, got %v", d)
	}
}

func TestParseCronExpr_FixedMinuteAndHour_Daily(t *testing.T) {
	d, err := ParseCronExpr("0 3 * * *")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != 24*time.Hour {
		t.Errorf("expected 24h, got %v", d)
	}
}

func TestParseCronExpr_InvalidFieldCount(t *testing.T) {
	_, err := ParseCronExpr("* * * *")
	if err == nil {
		t.Fatal("expected error for 4-field expression")
	}
}

func TestParseCronExpr_InvalidStepValue(t *testing.T) {
	_, err := ParseCronExpr("*/abc * * * *")
	if err == nil {
		t.Fatal("expected error for non-numeric step")
	}
}

func TestParseCronExpr_ZeroStep(t *testing.T) {
	_, err := ParseCronExpr("*/0 * * * *")
	if err == nil {
		t.Fatal("expected error for zero step")
	}
}

func TestParseCronExpr_UnsupportedExpression(t *testing.T) {
	_, err := ParseCronExpr("0 3 1 * *")
	if err == nil {
		t.Fatal("expected error for unsupported monthly expression")
	}
}
