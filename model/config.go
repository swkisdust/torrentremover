package model

import (
	"os"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Log      LogConfig         `json:"log"`
	Daemon   DaemonConfig      `json:"daemon"`
	Clients  map[string]Client `json:"clients,omitempty"`
	Profiles []Profile         `json:"profiles,omitempty"`
}

type Client struct {
	Type   string       `json:"type"`
	Config ClientConfig `json:"config"`
}

type ClientConfig struct {
	Host     string `json:"host"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`

	BasicUser string `json:"basic_user,omitempty"`
	BasicPass string `json:"basic_pass,omitempty"`

	InsecureTLS bool `json:"insecure_tls,omitempty"`
}

type LogConfig struct {
	Disabled bool   `json:"enabled"`
	Level    string `json:"level"`
}

type DaemonConfig struct {
	Enabled bool   `json:"enabled"`
	CronExp string `json:"cron_exp"`
}

func (c *Config) Read(f string) error {
	b, err := os.ReadFile(f)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(b, c)
}
