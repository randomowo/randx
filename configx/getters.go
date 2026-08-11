package configx

import (
	"time"

	"github.com/knadh/koanf/v2"
)

func (c *Config) Get(path string) any { return c.k.Get(path) }

func (c *Config) Exists(path string) bool { return c.k.Exists(path) }

func (c *Config) Keys() []string { return c.k.Keys() }

func (c *Config) KeyMap() koanf.KeyMap { return c.k.KeyMap() }

func (c *Config) MapKeys(path string) []string { return c.k.MapKeys(path) }

func (c *Config) All() map[string]any { return c.k.All() }

func (c *Config) Raw() map[string]any { return c.k.Raw() }

func (c *Config) Sprint() string { return c.k.Sprint() }

func (c *Config) Cut(path string) *Config { return &Config{k: c.k.Cut(path)} }

func (c *Config) Slices(path string) []*Config {
	cut := c.k.Slices(path)
	out := make([]*Config, 0, len(cut))
	for _, k := range cut {
		out = append(out, &Config{k: k})
	}

	return out
}

func (c *Config) String(path string) string { return c.k.String(path) }

func (c *Config) Strings(path string) []string { return c.k.Strings(path) }

func (c *Config) StringMap(path string) map[string]string { return c.k.StringMap(path) }

func (c *Config) StringsMap(path string) map[string][]string { return c.k.StringsMap(path) }

func (c *Config) Bytes(path string) []byte { return c.k.Bytes(path) }

func (c *Config) Bool(path string) bool { return c.k.Bool(path) }

func (c *Config) Bools(path string) []bool { return c.k.Bools(path) }

func (c *Config) BoolMap(path string) map[string]bool { return c.k.BoolMap(path) }

func (c *Config) Int(path string) int { return c.k.Int(path) }

func (c *Config) Ints(path string) []int { return c.k.Ints(path) }

func (c *Config) IntMap(path string) map[string]int { return c.k.IntMap(path) }

func (c *Config) Int64(path string) int64 { return c.k.Int64(path) }

func (c *Config) Int64s(path string) []int64 { return c.k.Int64s(path) }

func (c *Config) Int64Map(path string) map[string]int64 { return c.k.Int64Map(path) }

func (c *Config) Float64(path string) float64 { return c.k.Float64(path) }

func (c *Config) Float64s(path string) []float64 { return c.k.Float64s(path) }

func (c *Config) Float64Map(path string) map[string]float64 { return c.k.Float64Map(path) }

func (c *Config) Duration(path string) time.Duration { return c.k.Duration(path) }

func (c *Config) Time(path, layout string) time.Time { return c.k.Time(path, layout) }
