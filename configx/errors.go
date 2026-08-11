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
