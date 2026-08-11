package configx

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type leaf struct {
	path      string
	typ       string
	itemsType string
}

type defaultEntry struct {
	path  string
	value any
}

type schemaInfo struct {
	leaves   map[string]leaf
	envKeys  map[string]leaf
	defaults []defaultEntry
}

func parseSchema(raw []byte) (*schemaInfo, error) {
	var root map[string]any
	if err := json.Unmarshal(raw, &root); err != nil {
		return nil, fmt.Errorf("configx: parse schema: %w", err)
	}

	info := &schemaInfo{
		leaves:  map[string]leaf{},
		envKeys: map[string]leaf{},
	}
	walkSchema(root, "", info)
	sortDefaults(info.defaults)

	if err := buildEnvKeys(info); err != nil {
		return nil, err
	}

	return info, nil
}

func walkSchema(node map[string]any, path string, info *schemaInfo) {
	if path != "" {
		if def, ok := node["default"]; ok {
			info.defaults = append(info.defaults, defaultEntry{path: path, value: def})
		}
	}

	props, ok := node["properties"].(map[string]any)
	if !ok {
		if path != "" {
			info.leaves[path] = leaf{path: path, typ: nodeType(node), itemsType: itemsType(node)}
		}

		return
	}

	for name, child := range props {
		childPath := name
		if path != "" {
			childPath = path + "." + name
		}

		childNode, ok := child.(map[string]any)
		if !ok {
			info.leaves[childPath] = leaf{path: childPath, typ: "", itemsType: ""}

			continue
		}

		walkSchema(childNode, childPath, info)
	}
}

func nodeType(node map[string]any) string {
	t, _ := node["type"].(string)
	return t
}

func itemsType(node map[string]any) string {
	items, ok := node["items"].(map[string]any)
	if !ok {
		return ""
	}

	t, _ := items["type"].(string)

	return t
}

func envName(path string) string {
	return strings.ToUpper(strings.ReplaceAll(path, ".", "_"))
}

func sortDefaults(d []defaultEntry) {
	sort.Slice(d, func(i, j int) bool {
		di, dj := strings.Count(d[i].path, "."), strings.Count(d[j].path, ".")
		if di != dj {
			return di < dj
		}

		return d[i].path < d[j].path
	})
}

func buildEnvKeys(info *schemaInfo) error {
	byEnv := map[string][]string{}
	for path := range info.leaves {
		name := envName(path)
		byEnv[name] = append(byEnv[name], path)
	}

	collisions := map[string][]string{}
	for name, paths := range byEnv {
		if len(paths) > 1 {
			sort.Strings(paths)
			collisions[name] = paths

			continue
		}

		info.envKeys[name] = info.leaves[paths[0]]
	}

	if len(collisions) > 0 {
		return &EnvKeyCollisionError{Collisions: collisions}
	}

	return nil
}

func coerce(l leaf, v string) (any, error) {
	switch l.typ {
	case "integer":
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, &coerceError{typ: l.typ, value: v, err: err}
		}

		return n, nil

	case "number":
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, &coerceError{typ: l.typ, value: v, err: err}
		}

		return f, nil

	case "boolean":
		b, err := strconv.ParseBool(v)
		if err != nil {
			return nil, &coerceError{typ: l.typ, value: v, err: err}
		}

		return b, nil

	case "array":
		if v == "" {
			return []any{}, nil
		}

		parts := strings.Split(v, ",")
		out := make([]any, 0, len(parts))
		item := leaf{path: l.path, typ: l.itemsType}
		for _, p := range parts {
			c, err := coerce(item, p)
			if err != nil {
				return nil, err
			}
			out = append(out, c)
		}

		return out, nil

	default:
		return v, nil
	}
}
