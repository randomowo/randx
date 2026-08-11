package configx

import (
	"testing"
	"time"
)

const timestampSchema = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "additionalProperties": false,
  "properties": {
    "day": {"type": "string"},
    "quoted": {"type": "string"},
    "days": {"type": "array", "items": {"type": "string"}},
    "nested": {
      "type": "object",
      "additionalProperties": false,
      "properties": {"day": {"type": "string"}}
    }
  }
}`

func TestUnquotedDateBecomesRFC3339(t *testing.T) {
	c, err := New([]byte(timestampSchema), WithYAMLBytes([]byte("day: 2026-08-11\n")))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if got := c.String("day"); got != "2026-08-11T00:00:00Z" {
		t.Errorf("String = %q, want \"2026-08-11T00:00:00Z\"", got)
	}

	parsed := c.Time("day", time.RFC3339)
	if parsed.IsZero() {
		t.Fatal("Time returned the zero time")
	}
	if parsed.Year() != 2026 || parsed.Month() != time.August || parsed.Day() != 11 {
		t.Errorf("Time = %v, want 2026-08-11", parsed)
	}
}

func TestQuotedDateIsUntouched(t *testing.T) {
	c, err := New([]byte(timestampSchema), WithYAMLBytes([]byte(`quoted: "2026-08-11"`+"\n")))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if got := c.String("quoted"); got != "2026-08-11" {
		t.Errorf("String = %q, want \"2026-08-11\" unchanged", got)
	}
}

func TestNestedAndListedDatesNormalize(t *testing.T) {
	yamlSrc := "days:\n  - 2026-08-11\n  - 2026-08-12\nnested:\n  day: 2026-08-13\n"

	c, err := New([]byte(timestampSchema), WithYAMLBytes([]byte(yamlSrc)))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	days := c.Strings("days")
	if len(days) != 2 || days[0] != "2026-08-11T00:00:00Z" || days[1] != "2026-08-12T00:00:00Z" {
		t.Errorf("Strings = %v, want both dates as RFC3339", days)
	}

	if got := c.String("nested.day"); got != "2026-08-13T00:00:00Z" {
		t.Errorf("nested.day = %q, want \"2026-08-13T00:00:00Z\"", got)
	}
}

func TestEnvOverrideOfDateIsNotNormalized(t *testing.T) {
	t.Setenv("DAY", "whatever")

	c, err := New([]byte(timestampSchema), WithYAMLBytes([]byte("day: 2026-08-11\n")))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if got := c.String("day"); got != "whatever" {
		t.Errorf("day = %q, want \"whatever\" — env wins and there is no timestamp left to normalize", got)
	}
}

func TestConfigWithNoTimestampsIsUnaffected(t *testing.T) {
	c, err := New([]byte(timestampSchema), WithYAMLBytes([]byte("quoted: plain\n")))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if got := c.String("quoted"); got != "plain" {
		t.Errorf("quoted = %q, want \"plain\"", got)
	}
}
