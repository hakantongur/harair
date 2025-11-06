package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Registry struct {
	URL         string `yaml:"url"`
	APIURL      string `yaml:"api_url"`
	RegistryURL string `yaml:"registry_url"`
	Insecure    bool   `yaml:"insecure"`
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
}

type Config struct {
	SkopeoPath string              `yaml:"skopeo_path"` // "docker" or "skopeo"
	Registries map[string]Registry `yaml:"registries"`
	AuthStore  string              `yaml:"auth_store,omitempty"` // optional: where `login` persists creds
}

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Config
	if err := yaml.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	return &c, nil
}
