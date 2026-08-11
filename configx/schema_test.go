package configx

import (
	"reflect"
	"strings"
	"testing"
)

const testSchema = `{
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
            "debug": {"type": "boolean"},
            "ratio": {"type": "number"},
            "hosts": {"type": "array", "items": {"type": "string"}},
            "ports": {"type": "array", "items": {"type": "integer"}}
          }
        }
      }
    }
  }
}`

func TestParseSchemaLeaves(t *testing.T) {
	info, err := parseSchema([]byte(testSchema))
	if err != nil {
		t.Fatalf("parseSchema: %v", err)
	}

	want := map[string]leaf{
		"app.name":              {path: "app.name", typ: "string"},
		"app.log_level":         {path: "app.log_level", typ: "string"},
		"app.http.port":         {path: "app.http.port", typ: "integer"},
		"app.http.read-timeout": {path: "app.http.read-timeout", typ: "string"},
		"app.http.debug":        {path: "app.http.debug", typ: "boolean"},
		"app.http.ratio":        {path: "app.http.ratio", typ: "number"},
		"app.http.hosts":        {path: "app.http.hosts", typ: "array", itemsType: "string"},
		"app.http.ports":        {path: "app.http.ports", typ: "array", itemsType: "integer"},
	}

	if !reflect.DeepEqual(info.leaves, want) {
		t.Fatalf("leaves mismatch\n got: %#v\nwant: %#v", info.leaves, want)
	}
}

func TestParseSchemaUnionTypeIsUntyped(t *testing.T) {
	schema := `{"type":"object","properties":{"a":{"type":["string","null"]}}}`

	info, err := parseSchema([]byte(schema))
	if err != nil {
		t.Fatalf("parseSchema: %v", err)
	}

	if got := info.leaves["a"].typ; got != "" {
		t.Fatalf("typ = %q, want empty", got)
	}
}

func TestParseSchemaBooleanSubschemaIsLeaf(t *testing.T) {
	schema := `{"type":"object","properties":{"a":true,"b":{"type":"string"}}}`

	info, err := parseSchema([]byte(schema))
	if err != nil {
		t.Fatalf("parseSchema: %v", err)
	}

	if got := info.leaves["a"].typ; got != "" {
		t.Fatalf("a.typ = %q, want empty", got)
	}

	if got := info.leaves["b"].typ; got != "string" {
		t.Fatalf("b.typ = %q, want string", got)
	}
}

func TestParseSchemaInvalidJSON(t *testing.T) {
	_, err := parseSchema([]byte("{not json"))
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.HasPrefix(err.Error(), "configx:") {
		t.Fatalf("error prefix = %q, want configx:", err.Error())
	}
}

func TestEnvName(t *testing.T) {
	cases := map[string]string{
		"app.name":              "APP_NAME",
		"app.log_level":         "APP_LOG_LEVEL",
		"app.http.read-timeout": "APP_HTTP_READ-TIMEOUT",
	}

	for path, want := range cases {
		if got := envName(path); got != want {
			t.Errorf("envName(%q) = %q, want %q", path, got, want)
		}
	}
}
