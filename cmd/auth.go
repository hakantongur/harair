package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hakantongur/harair/internal/config"
)

type authMap map[string]struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func readAuthStore(p string) (authMap, error) {
	var m authMap
	if b, err := os.ReadFile(p); err == nil {
		if len(b) == 0 {
			return authMap{}, nil
		}
		if err := json.Unmarshal(b, &m); err != nil {
			return nil, fmt.Errorf("parse auth store: %w", err)
		}
		return m, nil
	}
	return authMap{}, nil
}

func writeAuthStore(p string, m authMap) error {
	if err := config.EnsureDir(p); err != nil {
		return err
	}
	b, _ := json.MarshalIndent(m, "", "  ")
	return os.WriteFile(p, b, 0o600)
}

func getCreds(cfg *config.Config, name string) (string, string, error) {
	m, err := readAuthStore(cfg.AuthStore)
	if err != nil {
		return "", "", err
	}
	c, ok := m[name]
	if !ok {
		return "", "", fmt.Errorf("no credentials for %s, run: harair login %s", name, name)
	}
	return c.Username, c.Password, nil
}

func authStorePathHomeFallback(p string) (string, error) {
	if filepath.IsAbs(p) {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, p), nil
}
