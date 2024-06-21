package config

type Config struct {
	Log    *Log    `yaml:"log"`
	Server *Server `yaml:"server"`
}
