package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Project Project           `yaml:"project" toml:"project"`
	Servers map[string]Server `yaml:"servers" toml:"servers"`
	Alert   Alert             `yaml:"alert" toml:"alert"`
}

type Project struct {
	Name string `yaml:"name" toml:"name"`
}

type Server struct {
	Host        string            `yaml:"host" toml:"host"`
	Credentials Credentials       `yaml:"credentials" toml:"credentials"`
	Webapps     map[string]Webapp `yaml:"webapps" toml:"webapps"`
	Docker      Docker            `yaml:"docker" toml:"docker"`
	SAR         SAR               `yaml:"sar" toml:"sar"`
}

type SAR struct {
	Enabled bool `yaml:"enabled" toml:"enabled"`
}

type Credentials struct {
	User     string `yaml:"user" toml:"user"`
	SSHKey   string `yaml:"ssh-key" toml:"ssh-key"`
	Password string `yaml:"password" toml:"password"`
}

type Webapp struct {
	URL               string `yaml:"url" toml:"url"`
	IgnoreCertificate bool   `yaml:"ignore-certificate" toml:"ignore-certificate"`
}

type Docker struct {
	Status bool `yaml:"status" toml:"status"`
}

type Alert struct {
	Telegram Telegram `yaml:"telegram" toml:"telegram"`
}

type Telegram struct {
	Enabled bool   `yaml:"enabled" toml:"enabled"`
	Mode    string `yaml:"mode" toml:"mode"`
	Token   string `yaml:"token" toml:"token"`
	ChatID  string `yaml:"chat-id" toml:"chat-id"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".yaml" || ext == ".yml" {
		err = yaml.Unmarshal(data, &config)
	} else if ext == ".toml" {
		err = toml.Unmarshal(data, &config)
	}

	if err != nil {
		return nil, err
	}

	return &config, nil
}
