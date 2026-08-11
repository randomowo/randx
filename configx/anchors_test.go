package configx

import (
	"strings"
	"testing"
)

const anchorSchema = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "additionalProperties": false,
  "properties": {
    "defaults": {
      "type": "object",
      "additionalProperties": false,
      "properties": {
        "host": {"type": "string"},
        "port": {"type": "integer"}
      }
    },
    "level": {"type": "string"},
    "hostlist": {"type": "array", "items": {"type": "string"}},
    "db": {
      "type": "object",
      "additionalProperties": false,
      "properties": {
        "host": {"type": "string"},
        "port": {"type": "integer"}
      }
    },
    "cache": {
      "type": "object",
      "additionalProperties": false,
      "properties": {
        "host": {"type": "string"},
        "port": {"type": "integer"}
      }
    },
    "log_level": {"type": "string"},
    "hosts": {"type": "array", "items": {"type": "string"}}
  }
}`

func TestYAMLAnchorsFromFile(t *testing.T) {
	c, err := New([]byte(anchorSchema), WithYAMLFile("testdata/anchors.yaml"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if got := c.String("db.host"); got != "localhost" {
		t.Errorf("db.host = %q, want \"localhost\" from the merge key", got)
	}
	if got := c.Int("db.port"); got != 6432 {
		t.Errorf("db.port = %d, want 6432 — a local key beats the merged one", got)
	}
	if got := c.Int("cache.port"); got != 5432 {
		t.Errorf("cache.port = %d, want 5432 from the merge key", got)
	}
	if got := c.String("log_level"); got != "info" {
		t.Errorf("log_level = %q, want \"info\" from the scalar alias", got)
	}

	hosts := c.Strings("hosts")
	if len(hosts) != 2 || hosts[0] != "a.example" || hosts[1] != "b.example" {
		t.Errorf("hosts = %v, want [a.example b.example] from the sequence alias", hosts)
	}
}

func TestYAMLMergeKeyNeverLeaks(t *testing.T) {
	c, err := New([]byte(anchorSchema), WithYAMLFile("testdata/anchors.yaml"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	for _, key := range c.Keys() {
		if strings.Contains(key, "<<") {
			t.Fatalf("merge key leaked into the config as %q", key)
		}
	}
}

func TestYAMLAnchorHolderMustBeDeclared(t *testing.T) {
	schema := `{
	  "type": "object",
	  "additionalProperties": false,
	  "properties": {
	    "db": {
	      "type": "object",
	      "additionalProperties": false,
	      "properties": {"port": {"type": "integer"}}
	    }
	  }
	}`

	yamlSrc := "defaults: &defaults\n  port: 5432\ndb:\n  <<: *defaults\n"

	_, err := New([]byte(schema), WithYAMLBytes([]byte(yamlSrc)))
	if err == nil {
		t.Fatal("expected the undeclared anchor holder to fail validation")
	}
	if !strings.Contains(err.Error(), "defaults") {
		t.Errorf("error should name the offending key, got: %v", err)
	}
}

func TestEnvOverridesMergedValue(t *testing.T) {
	t.Setenv("DB_HOST", "from-env")

	c, err := New([]byte(anchorSchema), WithYAMLFile("testdata/anchors.yaml"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if got := c.String("db.host"); got != "from-env" {
		t.Errorf("db.host = %q, want \"from-env\" — env outranks a merged value", got)
	}
	if got := c.String("cache.host"); got != "localhost" {
		t.Errorf("cache.host = %q, want \"localhost\" — untouched by the env var", got)
	}
}

func TestAnchorHolderGetsEnvKeys(t *testing.T) {
	t.Setenv("DEFAULTS_PORT", "1234")

	c, err := New([]byte(anchorSchema), WithYAMLFile("testdata/anchors.yaml"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if got := c.Int("defaults.port"); got != 1234 {
		t.Errorf("defaults.port = %d, want 1234 — a declared holder is an ordinary property", got)
	}
	if got := c.Int("db.port"); got != 6432 {
		t.Errorf("db.port = %d, want 6432 — anchors resolve at parse time, so env on the holder does not propagate", got)
	}
}
