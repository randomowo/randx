package configx

import "testing"

func TestPackageCompiles(t *testing.T) {
	c := &Config{}
	if c == nil {
		t.Fatal("expected a Config")
	}
}
