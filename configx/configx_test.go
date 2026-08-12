package configx

import (
	"errors"
	"strconv"
	"testing"
)

const newSchema = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "additionalProperties": false,
  "properties": {
    "app": {
      "type": "object",
      "additionalProperties": false,
      "properties": {
        "name": {"type": "string", "default": "randx"},
        "log_level": {"type": "string", "default": "info"},
        "http": {
          "type": "object",
          "additionalProperties": false,
          "properties": {
            "port": {"type": "integer", "default": 8080},
            "read-timeout": {"type": "string", "default": "5s"},
            "debug": {"type": "boolean", "default": false},
            "hosts": {"type": "array", "items": {"type": "string"}, "default": []}
          }
        }
      }
    }
  }
}`

func TestNewDefaultsOnly(t *testing.T) {
	c, err := New([]byte(newSchema))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if got := c.String("app.name"); got != "randx" {
		t.Errorf("app.name = %q, want \"randx\"", got)
	}
	if got := c.Int("app.http.port"); got != 8080 {
		t.Errorf("app.http.port = %d, want 8080", got)
	}
	if got := c.String("app.http.read-timeout"); got != "5s" {
		t.Errorf("app.http.read-timeout = %q, want \"5s\"", got)
	}
}

func TestNewYAMLOverridesDefaults(t *testing.T) {
	c, err := New([]byte(newSchema), WithYAMLBytes([]byte("app:\n  name: from-yaml\n")))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if got := c.String("app.name"); got != "from-yaml" {
		t.Errorf("app.name = %q, want \"from-yaml\"", got)
	}
	if got := c.Int("app.http.port"); got != 8080 {
		t.Errorf("app.http.port = %d, want the default 8080", got)
	}
}

func TestNewEnvOverridesYAML(t *testing.T) {
	t.Setenv("APP_NAME", "from-env")
	t.Setenv("APP_HTTP_PORT", "9999")

	c, err := New([]byte(newSchema), WithYAMLBytes([]byte("app:\n  name: from-yaml\n  http:\n    port: 7777\n")))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if got := c.String("app.name"); got != "from-env" {
		t.Errorf("app.name = %q, want \"from-env\"", got)
	}
	if got := c.Int("app.http.port"); got != 9999 {
		t.Errorf("app.http.port = %d, want 9999", got)
	}
}

func TestNewEnvHyphenAndUnderscoreKeys(t *testing.T) {
	t.Setenv("APP_HTTP_READ-TIMEOUT", "30s")
	t.Setenv("APP_LOG_LEVEL", "debug")

	c, err := New([]byte(newSchema))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if got := c.String("app.http.read-timeout"); got != "30s" {
		t.Errorf("app.http.read-timeout = %q, want \"30s\"", got)
	}
	if got := c.String("app.log_level"); got != "debug" {
		t.Errorf("app.log_level = %q, want \"debug\"", got)
	}
}

func TestNewEnvCoercesTypes(t *testing.T) {
	t.Setenv("APP_HTTP_DEBUG", "true")
	t.Setenv("APP_HTTP_HOSTS", "a.example,b.example")

	c, err := New([]byte(newSchema))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if !c.Bool("app.http.debug") {
		t.Error("app.http.debug = false, want true")
	}

	want := []string{"a.example", "b.example"}
	got := c.Strings("app.http.hosts")
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("app.http.hosts = %v, want %v", got, want)
	}
}

func TestNewIgnoresUnknownEnv(t *testing.T) {
	t.Setenv("TOTALLY_UNRELATED", "junk")

	c, err := New([]byte(newSchema))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if c.Exists("totally.unrelated") {
		t.Error("unknown env var leaked into the config")
	}
}

func TestNewEnvCoercionFailure(t *testing.T) {
	t.Setenv("APP_HTTP_PORT", "not-a-number")

	_, err := New([]byte(newSchema))
	if err == nil {
		t.Fatal("expected an error")
	}

	var valueErr *EnvValueError
	if !errors.As(err, &valueErr) {
		t.Fatalf("error is %T, want *EnvValueError: %v", err, err)
	}
	if valueErr.EnvKey != "APP_HTTP_PORT" {
		t.Errorf("EnvKey = %q, want \"APP_HTTP_PORT\"", valueErr.EnvKey)
	}
	if valueErr.Path != "app.http.port" {
		t.Errorf("Path = %q, want \"app.http.port\"", valueErr.Path)
	}
	if !errors.Is(err, strconv.ErrSyntax) {
		t.Error("expected the strconv error to be wrapped")
	}
}

func TestNewRejectsUnknownYAMLKey(t *testing.T) {
	_, err := New([]byte(newSchema), WithYAMLBytes([]byte("app:\n  nope: 1\n")))
	if err == nil {
		t.Fatal("expected additionalProperties:false to reject the key")
	}
}

func TestNewRejectsWrongType(t *testing.T) {
	_, err := New([]byte(newSchema), WithYAMLBytes([]byte("app:\n  http:\n    port: not-a-number\n")))
	if err == nil {
		t.Fatal("expected a validation error")
	}
}

func TestNewPropagatesCollisionError(t *testing.T) {
	schema := `{"type":"object","properties":{"a_b":{"type":"string"},"a":{"type":"object","properties":{"b":{"type":"string"}}}}}`

	_, err := New([]byte(schema))

	var collision *EnvKeyCollisionError
	if !errors.As(err, &collision) {
		t.Fatalf("error is %T, want *EnvKeyCollisionError: %v", err, err)
	}
}

func TestNewInvalidSchema(t *testing.T) {
	if _, err := New([]byte("{not json")); err == nil {
		t.Fatal("expected an error")
	}
}

func TestNewMissingYAMLFile(t *testing.T) {
	if _, err := New([]byte(newSchema), WithYAMLFile("testdata/does-not-exist.yaml")); err == nil {
		t.Fatal("expected an error")
	}
}

const envPrefixSchema = `{
  "type": "object",
  "additionalProperties": false,
  "properties": {
    "home": {"type": "string", "default": "schema-default"}
  }
}`

func TestWithEnvPrefixIgnoresUnprefixed(t *testing.T) {
	t.Setenv("HOME", "leaked-home")
	t.Setenv("APP_HOME", "prefixed-home")

	c, err := New([]byte(envPrefixSchema), WithEnvPrefix("APP_"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if got := c.String("home"); got != "prefixed-home" {
		t.Errorf("home = %q, want \"prefixed-home\"", got)
	}
}

func TestWithEnvPrefixStripsPrefix(t *testing.T) {
	t.Setenv("CFG_APP_HTTP_PORT", "9999")

	c, err := New([]byte(newSchema), WithEnvPrefix("CFG_"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if got := c.Int("app.http.port"); got != 9999 {
		t.Errorf("app.http.port = %d, want 9999", got)
	}
}

func TestNewEmptyEnvKeepsYAML(t *testing.T) {
	t.Setenv("APP_NAME", "")

	c, err := New([]byte(newSchema), WithYAMLBytes([]byte("app:\n  name: from-yaml\n")))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if got := c.String("app.name"); got != "from-yaml" {
		t.Errorf("app.name = %q, want \"from-yaml\"", got)
	}
}

func TestNewEmptyEnvKeepsDefault(t *testing.T) {
	t.Setenv("APP_HTTP_PORT", "")

	c, err := New([]byte(newSchema))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if got := c.Int("app.http.port"); got != 8080 {
		t.Errorf("app.http.port = %d, want the default 8080", got)
	}
}

func TestNewEmptyEnvStillReadsSetValues(t *testing.T) {
	t.Setenv("APP_NAME", "from-env")
	t.Setenv("APP_HTTP_PORT", "")

	c, err := New([]byte(newSchema))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if got := c.String("app.name"); got != "from-env" {
		t.Errorf("app.name = %q, want \"from-env\"", got)
	}
}

func TestNewAppliesSourcesInOrder(t *testing.T) {
	c, err := New([]byte(newSchema),
		WithYAMLBytes([]byte("app:\n  name: first\n")),
		WithYAMLBytes([]byte("app:\n  name: second\n")),
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if got := c.String("app.name"); got != "second" {
		t.Errorf("app.name = %q, want \"second\"", got)
	}
}

const shallowDeepSchema = `{
  "type": "object",
  "properties": {
    "a": {
      "type": "object",
      "default": {"b": {"c": "shallow", "d": "only-shallow"}},
      "properties": {
        "b": {
          "type": "object",
          "properties": {
            "c": {"type": "string", "default": "deep"}
          }
        }
      }
    }
  }
}`

func TestNewAppliesShallowDefaultsBeforeDeep(t *testing.T) {
	c, err := New([]byte(shallowDeepSchema))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if got := c.String("a.b.c"); got != "deep" {
		t.Errorf("a.b.c = %q, want \"deep\"", got)
	}
	if got := c.String("a.b.d"); got != "only-shallow" {
		t.Errorf("a.b.d = %q, want \"only-shallow\"", got)
	}
}
