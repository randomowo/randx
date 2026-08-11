package configx

import (
	"errors"
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

func TestParseSchemaDefaults(t *testing.T) {
	info, err := parseSchema([]byte(testSchema))
	if err != nil {
		t.Fatalf("parseSchema: %v", err)
	}

	want := []defaultEntry{
		{path: "app.log_level", value: "info"},
		{path: "app.name", value: "randx"},
		{path: "app.http.port", value: float64(8080)},
		{path: "app.http.read-timeout", value: "5s"},
	}

	if !reflect.DeepEqual(info.defaults, want) {
		t.Fatalf("defaults mismatch\n got: %#v\nwant: %#v", info.defaults, want)
	}
}

func TestParseSchemaDefaultsShallowFirst(t *testing.T) {
	schema := `{
	  "type": "object",
	  "properties": {
	    "a": {
	      "type": "object",
	      "default": {"b": {"c": "shallow"}},
	      "properties": {
	        "b": {
	          "type": "object",
	          "properties": {"c": {"type": "string", "default": "deep"}}
	        }
	      }
	    }
	  }
	}`

	info, err := parseSchema([]byte(schema))
	if err != nil {
		t.Fatalf("parseSchema: %v", err)
	}

	if len(info.defaults) != 2 {
		t.Fatalf("got %d defaults, want 2", len(info.defaults))
	}
	if info.defaults[0].path != "a" {
		t.Errorf("defaults[0].path = %q, want \"a\"", info.defaults[0].path)
	}
	if info.defaults[1].path != "a.b.c" {
		t.Errorf("defaults[1].path = %q, want \"a.b.c\"", info.defaults[1].path)
	}
}

func TestParseSchemaIgnoresRootDefault(t *testing.T) {
	schema := `{"type":"object","default":{"a":1},"properties":{"a":{"type":"integer"}}}`

	info, err := parseSchema([]byte(schema))
	if err != nil {
		t.Fatalf("parseSchema: %v", err)
	}

	if len(info.defaults) != 0 {
		t.Fatalf("got %d defaults, want 0", len(info.defaults))
	}
}

func TestParseSchemaEnvKeys(t *testing.T) {
	info, err := parseSchema([]byte(testSchema))
	if err != nil {
		t.Fatalf("parseSchema: %v", err)
	}

	cases := map[string]string{
		"APP_NAME":              "app.name",
		"APP_LOG_LEVEL":         "app.log_level",
		"APP_HTTP_PORT":         "app.http.port",
		"APP_HTTP_READ-TIMEOUT": "app.http.read-timeout",
		"APP_HTTP_PORTS":        "app.http.ports",
	}

	for env, wantPath := range cases {
		lf, ok := info.envKeys[env]
		if !ok {
			t.Errorf("envKeys missing %q", env)
			continue
		}
		if lf.path != wantPath {
			t.Errorf("envKeys[%q].path = %q, want %q", env, lf.path, wantPath)
		}
	}
}

func TestParseSchemaEnvKeyCollision(t *testing.T) {
	schema := `{
	  "type": "object",
	  "properties": {
	    "a": {
	      "type": "object",
	      "properties": {
	        "b_c": {"type": "string"},
	        "b": {"type": "object", "properties": {"c": {"type": "string"}}}
	      }
	    }
	  }
	}`

	_, err := parseSchema([]byte(schema))
	if err == nil {
		t.Fatal("expected a collision error")
	}

	var collision *EnvKeyCollisionError
	if !errors.As(err, &collision) {
		t.Fatalf("error is %T, want *EnvKeyCollisionError", err)
	}

	want := `configx: env key collision: A_B_C is derived from keys "a.b.c", "a.b_c"`
	if err.Error() != want {
		t.Fatalf("Error() = %q, want %q", err.Error(), want)
	}
}

func TestParseSchemaReportsEveryCollision(t *testing.T) {
	schema := `{
	  "type": "object",
	  "properties": {
	    "a_b": {"type": "string"},
	    "a": {"type": "object", "properties": {"b": {"type": "string"}}},
	    "x_y": {"type": "string"},
	    "x": {"type": "object", "properties": {"y": {"type": "string"}}}
	  }
	}`

	_, err := parseSchema([]byte(schema))
	if err == nil {
		t.Fatal("expected a collision error")
	}

	want := `configx: env key collision: A_B is derived from keys "a.b", "a_b"; X_Y is derived from keys "x.y", "x_y"`
	if err.Error() != want {
		t.Fatalf("Error() = %q, want %q", err.Error(), want)
	}
}
