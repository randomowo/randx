package configx

import (
	"testing"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
)

func applyOptions(t *testing.T, opts ...Option) *koanf.Koanf {
	t.Helper()

	l := &loader{}
	for _, opt := range opts {
		if err := opt(l); err != nil {
			t.Fatalf("option: %v", err)
		}
	}

	k := koanf.New(".")
	for _, s := range l.sources {
		if err := k.Load(s.provider, s.parser); err != nil {
			t.Fatalf("load: %v", err)
		}
	}

	return k
}

func TestWithYAMLFile(t *testing.T) {
	k := applyOptions(t, WithYAMLFile("testdata/config.yaml"))

	if got := k.String("app.name"); got != "from-file" {
		t.Errorf("app.name = %q, want \"from-file\"", got)
	}
	if got := k.Int("app.http.port"); got != 9090 {
		t.Errorf("app.http.port = %d, want 9090", got)
	}
}

func TestWithYAMLBytes(t *testing.T) {
	k := applyOptions(t, WithYAMLBytes([]byte("app:\n  name: from-bytes\n")))

	if got := k.String("app.name"); got != "from-bytes" {
		t.Errorf("app.name = %q, want \"from-bytes\"", got)
	}
}

func TestWithProvider(t *testing.T) {
	k := applyOptions(t, WithProvider(rawbytes.Provider([]byte("app:\n  name: from-provider\n")), yaml.Parser()))

	if got := k.String("app.name"); got != "from-provider" {
		t.Errorf("app.name = %q, want \"from-provider\"", got)
	}
}

func TestOptionsApplyInOrder(t *testing.T) {
	k := applyOptions(t,
		WithYAMLBytes([]byte("app:\n  name: first\n")),
		WithYAMLBytes([]byte("app:\n  name: second\n")),
	)

	if got := k.String("app.name"); got != "second" {
		t.Errorf("app.name = %q, want \"second\"", got)
	}
}

func TestWithYAMLFileMissing(t *testing.T) {
	l := &loader{}
	if err := WithYAMLFile("testdata/does-not-exist.yaml")(l); err != nil {
		t.Fatalf("option itself should not fail: %v", err)
	}

	k := koanf.New(".")
	if err := k.Load(l.sources[0].provider, l.sources[0].parser); err == nil {
		t.Fatal("expected loading a missing file to fail")
	}
}
