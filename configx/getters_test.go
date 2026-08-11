package configx

import (
	"testing"
	"time"
)

const gettersSchema = `{
  "type": "object",
  "additionalProperties": false,
  "properties": {
    "s": {"type": "string"},
    "i": {"type": "integer"},
    "f": {"type": "number"},
    "b": {"type": "boolean"},
    "d": {"type": "string"},
    "t": {"type": "string"},
    "ss": {"type": "array", "items": {"type": "string"}},
    "is": {"type": "array", "items": {"type": "integer"}},
    "fs": {"type": "array", "items": {"type": "number"}},
    "bs": {"type": "array", "items": {"type": "boolean"}},
    "m": {"type": "object", "additionalProperties": {"type": "string"}},
    "im": {"type": "object", "additionalProperties": {"type": "integer"}},
    "fm": {"type": "object", "additionalProperties": {"type": "number"}},
    "bm": {"type": "object", "additionalProperties": {"type": "boolean"}},
    "sm": {"type": "object", "additionalProperties": {"type": "array", "items": {"type": "string"}}},
    "nested": {
      "type": "object",
      "additionalProperties": false,
      "properties": {"x": {"type": "string"}}
    }
  }
}`

const gettersYAML = `
s: hello
i: 42
f: 1.5
b: true
d: 30s
t: "2026-08-11"
ss: [a, b]
is: [1, 2]
fs: [1.5, 2.5]
bs: [true, false]
m:
  k: v
im:
  k: 7
fm:
  k: 1.5
bm:
  k: true
sm:
  k: [x, y]
nested:
  x: deep
`

func newGettersConfig(t *testing.T) *Config {
	t.Helper()

	c, err := New([]byte(gettersSchema), WithYAMLBytes([]byte(gettersYAML)))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	return c
}

func TestScalarGetters(t *testing.T) {
	c := newGettersConfig(t)

	if got := c.String("s"); got != "hello" {
		t.Errorf("String = %q, want \"hello\"", got)
	}
	if got := c.Int("i"); got != 42 {
		t.Errorf("Int = %d, want 42", got)
	}
	if got := c.Int64("i"); got != 42 {
		t.Errorf("Int64 = %d, want 42", got)
	}
	if got := c.Float64("f"); got != 1.5 {
		t.Errorf("Float64 = %v, want 1.5", got)
	}
	if !c.Bool("b") {
		t.Error("Bool = false, want true")
	}
	if got := c.Duration("d"); got != 30*time.Second {
		t.Errorf("Duration = %v, want 30s", got)
	}
	if got := c.Time("t", time.DateOnly); got.Year() != 2026 {
		t.Errorf("Time year = %d, want 2026", got.Year())
	}
	if got := c.Bytes("s"); string(got) != "hello" {
		t.Errorf("Bytes = %q, want \"hello\"", got)
	}
	if got := c.Get("s"); got != "hello" {
		t.Errorf("Get = %v, want \"hello\"", got)
	}
}

func TestSliceGetters(t *testing.T) {
	c := newGettersConfig(t)

	if got := c.Strings("ss"); len(got) != 2 || got[0] != "a" {
		t.Errorf("Strings = %v, want [a b]", got)
	}
	if got := c.Ints("is"); len(got) != 2 || got[1] != 2 {
		t.Errorf("Ints = %v, want [1 2]", got)
	}
	if got := c.Int64s("is"); len(got) != 2 || got[1] != 2 {
		t.Errorf("Int64s = %v, want [1 2]", got)
	}
	if got := c.Float64s("fs"); len(got) != 2 || got[0] != 1.5 {
		t.Errorf("Float64s = %v, want [1.5 2.5]", got)
	}
	if got := c.Bools("bs"); len(got) != 2 || !got[0] {
		t.Errorf("Bools = %v, want [true false]", got)
	}
}

func TestMapGetters(t *testing.T) {
	c := newGettersConfig(t)

	if got := c.StringMap("m"); got["k"] != "v" {
		t.Errorf("StringMap = %v, want map[k:v]", got)
	}
	if got := c.MapKeys("m"); len(got) != 1 || got[0] != "k" {
		t.Errorf("MapKeys = %v, want [k]", got)
	}
	if got := c.StringsMap("sm"); len(got["k"]) != 2 || got["k"][0] != "x" {
		t.Errorf("StringsMap = %v, want map[k:[x y]]", got)
	}
	if got := c.IntMap("im"); got["k"] != 7 {
		t.Errorf("IntMap = %v, want map[k:7]", got)
	}
	if got := c.Int64Map("im"); got["k"] != 7 {
		t.Errorf("Int64Map = %v, want map[k:7]", got)
	}
	if got := c.Float64Map("fm"); got["k"] != 1.5 {
		t.Errorf("Float64Map = %v, want map[k:1.5]", got)
	}
	if got := c.BoolMap("bm"); !got["k"] {
		t.Errorf("BoolMap = %v, want map[k:true]", got)
	}
}

func TestStructureGetters(t *testing.T) {
	c := newGettersConfig(t)

	if !c.Exists("nested.x") {
		t.Error("Exists(nested.x) = false, want true")
	}
	if c.Exists("nope") {
		t.Error("Exists(nope) = true, want false")
	}
	if len(c.Keys()) == 0 {
		t.Error("Keys is empty")
	}
	if len(c.KeyMap()) == 0 {
		t.Error("KeyMap is empty")
	}
	if len(c.All()) == 0 {
		t.Error("All is empty")
	}
	if len(c.Raw()) == 0 {
		t.Error("Raw is empty")
	}
	if c.Sprint() == "" {
		t.Error("Sprint is empty")
	}
}

func TestCutReturnsConfig(t *testing.T) {
	c := newGettersConfig(t)

	cut := c.Cut("nested")
	if got := cut.String("x"); got != "deep" {
		t.Errorf("Cut(nested).String(x) = %q, want \"deep\"", got)
	}
}

func TestSlicesReturnsConfigs(t *testing.T) {
	schema := `{
	  "type": "object",
	  "additionalProperties": false,
	  "properties": {
	    "items": {
	      "type": "array",
	      "items": {
	        "type": "object",
	        "additionalProperties": false,
	        "properties": {"name": {"type": "string"}}
	      }
	    }
	  }
	}`

	c, err := New([]byte(schema), WithYAMLBytes([]byte("items:\n  - name: one\n  - name: two\n")))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	slices := c.Slices("items")
	if len(slices) != 2 {
		t.Fatalf("got %d slices, want 2", len(slices))
	}
	if got := slices[0].String("name"); got != "one" {
		t.Errorf("slices[0].String(name) = %q, want \"one\"", got)
	}
}
