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

func getCreds(cfg *config.Config, name string) (string, string, bool) {
	r, ok := cfg.Registries[name]
	if !ok {
		return "", "", false
	}
	return r.Username, r.Password, true
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
