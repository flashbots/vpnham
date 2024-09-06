package config

type Config struct {
	Version string `yaml:"-"`

	Log    *Log    `yaml:"log"`
	Server *Server `yaml:"server"`
}
