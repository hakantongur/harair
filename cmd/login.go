package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/hakantongur/harair/internal/config"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login [registry]",
	Short: "Store credentials for a registry (writes to auth store file)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		regName := args[0]

		cfg, err := config.Load(cfgPath)
		if err != nil {
			return err
		}
		reg, ok := cfg.Registries[regName]
		if !ok {
			return fmt.Errorf("unknown registry %q in %s", regName, cfgPath)
		}

		// credentials must be defined in config.yaml (or you could prompt here)
		if reg.Username == "" || reg.Password == "" {
			return errors.New("username/password not set in config.yaml for this registry")
		}

		// pick auth store path
		authPath := cfg.AuthStore
		if authPath == "" {
			// default to user home (cross-platform enough for now)
			home, _ := os.UserHomeDir()
			authPath = filepath.Join(home, ".harair", "auth.json")
		}

		if err := writeAuthStore(authPath, regName, reg.Username, reg.Password); err != nil {
			return err
		}

		color.Green("Saved credentials for %s to %s", regName, authPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}

// simple file format: { "<registryName>": { "username": "...", "password": "..." }, ... }
type authEntry struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type authStore map[string]authEntry

func writeAuthStore(path, regName, user, pass string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var store authStore
	b, err := os.ReadFile(path)
	if err == nil {
		_ = json.Unmarshal(b, &store)
	}
	if store == nil {
		store = make(authStore)
	}
	store[regName] = authEntry{Username: user, Password: pass}

	out, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o600)
}
