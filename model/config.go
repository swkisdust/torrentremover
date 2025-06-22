package model

import (
	"fmt"
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
	Type   string         `json:"type"`
	Config map[string]any `json:"config"`
}

type LogConfig struct {
	Disabled bool   `json:"enabled"`
	Level    string `json:"level"`
}

type DaemonConfig struct {
	Disabled bool   `json:"disabled"`
	CronExp  string `json:"cron_exp"`
}

func (c *Config) Read(f string) error {
	b, err := os.ReadFile(f)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(b, c); err != nil {
		return err
	}

	for i, profile := range c.Profiles {
		for j, st := range profile.Strategy {
			if st.Name == "" {
				return fmt.Errorf("profiles[%d].strategy[%d] needs a name", i, j)
			}
		}
	}

	return nil
}
