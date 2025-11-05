package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/hakantongur/harair/internal/config"
	"github.com/hakantongur/harair/internal/harbor"
	"github.com/spf13/cobra"
)

var (
	lsProject string
	lsRepo    string
)

var lsCmd = &cobra.Command{
	Use:   "ls [registry]",
	Short: "List repos or tags from a registry (via Harbor API)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		reg := args[0]

		cfg, err := config.Load(cfgPath)
		if err != nil {
			return err
		}
		r, ok := cfg.Registries[reg]
		if !ok {
			return fmt.Errorf("registry %q not found in %s", reg, cfgPath)
		}
		user, pass, err := getCreds(cfg, reg)
		if err != nil {
			// Mocks don’t require auth—fall back to blanks if no creds stored yet
			user, pass = "", ""
		}
		apiBase := r.APIURL
		if apiBase == "" {
			apiBase = r.URL
		}

		hc := harbor.New(apiBase, user, pass, r.Insecure)

		if lsProject == "" {
			color.Yellow("Tip: use --project <name> (and optionally --repo <name>)")
			return nil
		}

		if lsRepo == "" {
			repos, err := hc.ListRepos(lsProject)
			if err != nil {
				return err
			}
			color.Cyan("%s/%s repos:", reg, lsProject)
			for _, rr := range repos {
				// API returns "<project>/<repo>"
				name := rr.Name
				if parts := strings.SplitN(rr.Name, "/", 2); len(parts) == 2 {
					name = parts[1]
				}
				fmt.Println("-", name)
			}
			return nil
		}

		arts, err := hc.ListArtifacts(lsProject, lsRepo)
		if err != nil {
			return err
		}
		color.Cyan("%s/%s/%s artifacts:", reg, lsProject, lsRepo)
		for _, a := range arts {
			var tags []string
			for _, t := range a.Tags {
				tags = append(tags, t.Name)
			}
			fmt.Printf("digest=%s tags=%v\n", a.Digest, tags)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(lsCmd)
	lsCmd.Flags().StringVar(&lsProject, "project", "", "Project to list")
	lsCmd.Flags().StringVar(&lsRepo, "repo", "", "Repo to list tags for")
}
