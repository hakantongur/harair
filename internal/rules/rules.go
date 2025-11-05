package rules

import (
	"os"

	"gopkg.in/yaml.v3"
)

type File struct {
	Projects []Project `yaml:"projects"`
}

type Project struct {
	Name     string   `yaml:"name"`
	Includes []string `yaml:"includes"` // repo name globs to include (empty => include all)
	Excludes []string `yaml:"excludes"` // repo name globs to exclude
	Tags     []string `yaml:"tags"`     // tag globs (empty => all)
}

type ImageInclude struct {
	Type    string   `yaml:"type"` // "image"
	From    string   `yaml:"from"`
	Project string   `yaml:"project"`
	Repo    string   `yaml:"repo"`
	Tags    []string `yaml:"tags"`
}

type HelmInclude struct {
	Type     string   `yaml:"type"` // "helm"
	From     string   `yaml:"from"`
	Project  string   `yaml:"project"`
	Name     string   `yaml:"name"`
	Versions []string `yaml:"versions"`
}

type RuleSet struct {
	Include []map[string]any `yaml:"include"`
	Exclude []map[string]any `yaml:"exclude"`
}

func Load(path string) (*File, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var f File
	if err := yaml.Unmarshal(b, &f); err != nil {
		return nil, err
	}
	return &f, nil
}
