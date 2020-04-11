package cfg

import (
	"gopkg.in/yaml.v2"
	"io"
)

type Config struct {
	Features map[string]Feature `yaml:"features"`
}

type Feature struct {
	Branch string `yaml:"branch"`
}

func FromYaml(file io.Reader) (config *Config, err error) {
	config = new(Config)
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(config)
	return
}
