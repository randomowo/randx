package configx

import (
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
)

type source struct {
	provider koanf.Provider
	parser   koanf.Parser
}

type loader struct {
	sources []source
}

type Option func(*loader) error

func WithYAMLFile(path string) Option {
	return func(l *loader) error {
		l.sources = append(l.sources, source{provider: file.Provider(path), parser: yaml.Parser()})

		return nil
	}
}

func WithYAMLBytes(b []byte) Option {
	return func(l *loader) error {
		l.sources = append(l.sources, source{provider: rawbytes.Provider(b), parser: yaml.Parser()})

		return nil
	}
}

func WithProvider(p koanf.Provider, parser koanf.Parser) Option {
	return func(l *loader) error {
		l.sources = append(l.sources, source{provider: p, parser: parser})

		return nil
	}
}
