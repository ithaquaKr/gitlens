package config

type Config struct{}
type LoadOptions struct {
	ExplicitPath string
	Overrides    Overrides
}
type Overrides struct{ Provider, APIKey, Model, Theme string }

func Load(opts LoadOptions) (*Config, error) { return &Config{}, nil }
