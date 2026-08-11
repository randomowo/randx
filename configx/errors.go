package configx

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type EnvKeyCollisionError struct {
	Collisions map[string][]string
}

func (e *EnvKeyCollisionError) Error() string {
	names := make([]string, 0, len(e.Collisions))
	for name := range e.Collisions {
		names = append(names, name)
	}
	sort.Strings(names)

	parts := make([]string, 0, len(names))
	for _, name := range names {
		paths := e.Collisions[name]
		quoted := make([]string, 0, len(paths))
		for _, p := range paths {
			quoted = append(quoted, strconv.Quote(p))
		}
		parts = append(parts, fmt.Sprintf("%s is derived from keys %s", name, strings.Join(quoted, ", ")))
	}

	return "configx: env key collision: " + strings.Join(parts, "; ")
}

type coerceError struct {
	typ   string
	value string
	err   error
}

func (e *coerceError) Error() string {
	return fmt.Sprintf("cannot parse %q as %s", e.value, e.typ)
}

func (e *coerceError) Unwrap() error { return e.err }

type EnvValueError struct {
	EnvKey string
	Path   string
	Type   string
	Value  string
	Err    error
}

func (e *EnvValueError) Error() string {
	return fmt.Sprintf("configx: env %s for key %q: cannot parse %q as %s", e.EnvKey, e.Path, e.Value, e.Type)
}

func (e *EnvValueError) Unwrap() error { return e.Err }
