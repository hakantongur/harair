package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/hakantongur/harair/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var loginCmd = &cobra.Command{
	Use:   "login [registry]",
	Short: "Store credentials for a registry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		registry := args[0]

		cfg, err := config.Load(cfgPath)
		if err != nil {
			return err
		}
		// expand auth path to absolute (e.g., ~/.harair/auth.json)
		authPath, err := authStorePathHomeFallback(cfg.AuthStore)
		if err != nil {
			return err
		}

		// ensure the registry exists in config.yaml
		if _, ok := cfg.Registries[registry]; !ok {
			return fmt.Errorf("registry %q not found in %s", registry, cfgPath)
		}

		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Username: ")
		userRaw, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		username := strings.TrimSpace(userRaw)

		fmt.Print("Password: ")
		// disable echo for password
		bytePass, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			return err
		}
		password := strings.TrimSpace(string(bytePass))

		// read existing store, update, write back
		m, err := readAuthStore(authPath)
		if err != nil {
			return err
		}
		m[registry] = struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}{username, password}

		if err := writeAuthStore(authPath, m); err != nil {
			return err
		}
		color.Green("Saved credentials for %s at %s", registry, authPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
