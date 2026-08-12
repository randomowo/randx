package configx

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/knadh/koanf/providers/confmap"
	env "github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/v2"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

const schemaResourceURL = "mem://configx/schema.json"

type Config struct {
	k *koanf.Koanf
}

func New(schema []byte, opts ...Option) (*Config, error) {
	info, err := parseSchema(schema)
	if err != nil {
		return nil, err
	}

	compiled, err := compileSchema(schema)
	if err != nil {
		return nil, err
	}

	l := &loader{}
	for _, opt := range opts {
		if err := opt(l); err != nil {
			return nil, err
		}
	}

	k := koanf.New(".")

	for _, d := range info.defaults {
		if err := k.Load(confmap.Provider(map[string]any{d.path: d.value}, "."), nil); err != nil {
			return nil, fmt.Errorf("configx: load default %q: %w", d.path, err)
		}
	}

	for i, s := range l.sources {
		if err := k.Load(s.provider, s.parser); err != nil {
			return nil, fmt.Errorf("configx: load source %d: %w", i, err)
		}
	}

	var envErrs []*EnvValueError
	provider := env.Provider(".", env.Opt{
		Prefix: l.envPrefix,
		TransformFunc: func(name, value string) (string, any) {
			// A process manager passes an unset optional variable as an empty
			// string, which should leave a source or a default alone.
			if value == "" {
				return "", nil
			}

			lf, ok := info.envKeys[strings.TrimPrefix(name, l.envPrefix)]
			if !ok {
				return "", nil
			}

			coerced, err := coerce(lf, value)
			if err != nil {
				envErrs = append(envErrs, &EnvValueError{
					EnvKey: name,
					Path:   lf.path,
					Err:    err,
				})

				return "", nil
			}

			return lf.path, coerced
		},
	})

	if err := k.Load(provider, nil); err != nil {
		return nil, fmt.Errorf("configx: load env: %w", err)
	}
	if len(envErrs) > 0 {
		sort.Slice(envErrs, func(i, j int) bool { return envErrs[i].EnvKey < envErrs[j].EnvKey })

		joined := make([]error, len(envErrs))
		for i, e := range envErrs {
			joined[i] = e
		}

		return nil, errors.Join(joined...)
	}

	if err := normalizeTimes(k); err != nil {
		return nil, err
	}

	if err := validate(compiled, k); err != nil {
		return nil, err
	}

	return &Config{k: k}, nil
}

func compileSchema(raw []byte) (*jsonschema.Schema, error) {
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("configx: parse schema: %w", err)
	}

	c := jsonschema.NewCompiler()
	if err := c.AddResource(schemaResourceURL, doc); err != nil {
		return nil, fmt.Errorf("configx: compile schema: %w", err)
	}

	sch, err := c.Compile(schemaResourceURL)
	if err != nil {
		return nil, fmt.Errorf("configx: compile schema: %w", err)
	}

	return sch, nil
}

func normalizeTimes(k *koanf.Koanf) error {
	replacements := map[string]any{}
	for path, value := range k.All() {
		normalized, changed := normalizeValue(value)
		if changed {
			replacements[path] = normalized
		}
	}

	if len(replacements) == 0 {
		return nil
	}

	if err := k.Load(confmap.Provider(replacements, "."), nil); err != nil {
		return fmt.Errorf("configx: normalize timestamps: %w", err)
	}

	return nil
}

func normalizeValue(v any) (any, bool) {
	switch t := v.(type) {
	case time.Time:
		return t.Format(time.RFC3339Nano), true

	case []any:
		out := make([]any, len(t))
		changed := false
		for i, item := range t {
			normalized, itemChanged := normalizeValue(item)
			out[i] = normalized
			changed = changed || itemChanged
		}

		return out, changed

	case map[string]any:
		out := make(map[string]any, len(t))
		changed := false
		for key, item := range t {
			normalized, itemChanged := normalizeValue(item)
			out[key] = normalized
			changed = changed || itemChanged
		}

		return out, changed

	default:
		return v, false
	}
}

func validate(sch *jsonschema.Schema, k *koanf.Koanf) error {
	b, err := json.Marshal(k.Raw())
	if err != nil {
		return fmt.Errorf("configx: encode config: %w", err)
	}

	inst, err := jsonschema.UnmarshalJSON(bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("configx: decode config: %w", err)
	}

	if err := sch.Validate(inst); err != nil {
		return fmt.Errorf("configx: config invalid: %w", err)
	}

	return nil
}
