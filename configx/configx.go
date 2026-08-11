package configx

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

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

	for _, s := range l.sources {
		if err := k.Load(s.provider, s.parser); err != nil {
			return nil, fmt.Errorf("configx: load source: %w", err)
		}
	}

	var envErrs []error
	provider := env.Provider(".", env.Opt{
		TransformFunc: func(name, value string) (string, any) {
			lf, ok := info.envKeys[name]
			if !ok {
				return "", nil
			}

			coerced, err := coerce(lf, value)
			if err != nil {
				var ce *coerceError
				errors.As(err, &ce)
				envErrs = append(envErrs, &EnvValueError{
					EnvKey: name,
					Path:   lf.path,
					Type:   ce.typ,
					Value:  ce.value,
					Err:    ce.err,
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
		return nil, errors.Join(envErrs...)
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
