package cmd

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/hakantongur/harair/internal/config"
	"github.com/hakantongur/harair/internal/harbor"
	"github.com/hakantongur/harair/internal/rules"
	"github.com/hakantongur/harair/internal/shell"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

// ----- flags -----
var (
	syncProject       string
	syncRepo          string
	syncTags          []string
	dryRun            bool
	syncDockerNetwork string
	syncRulesPath     string
	maxConcurrent     int
)

// ----- types -----
type copyTask struct {
	srcRef string
	dstRef string
}

// ----- command -----
var syncCmd = &cobra.Command{
	Use:   "sync [from-registry] [to-registry]",
	Short: "Copy images between registries (defaults to --dry-run)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		fromReg, toReg := args[0], args[1]

		if syncProject == "" {
			color.Red("Please provide --project")
			return nil
		}

		cfg, err := config.Load(cfgPath)
		if err != nil {
			return err
		}
		fr, ok := cfg.Registries[fromReg]
		if !ok {
			return fmt.Errorf("registry %q not in %s", fromReg, cfgPath)
		}
		tr, ok := cfg.Registries[toReg]
		if !ok {
			return fmt.Errorf("registry %q not in %s", toReg, cfgPath)
		}

		fu, fp, _ := getCreds(cfg, fromReg) // optional (mocks may not need)
		tu, tp, _ := getCreds(cfg, toReg)

		// Harbor API bases (discovery)
		srcAPI := fr.APIURL
		if srcAPI == "" {
			srcAPI = fr.URL // legacy fallback
		}
		if strings.TrimSpace(srcAPI) == "" {
			return fmt.Errorf("registry %q: api_url/url is empty in config.yaml", fromReg)
		}
		srcHC := harbor.New(srcAPI, fu, fp, fr.Insecure)

		// (currently we don't need dst API for discovery, keep for future)
		// dstAPI := tr.APIURL
		// if dstAPI == "" { dstAPI = tr.URL }

		// Build a plan: list of {repo, tag globs}
		type planItem struct {
			Repo string
			Tags []string
		}
		var plan []planItem

		if syncRulesPath != "" {
			rs, err := rules.Load(syncRulesPath)
			if err != nil {
				return fmt.Errorf("load rules: %w", err)
			}
			for _, p := range rs.Projects {
				if p.Name != syncProject {
					continue
				}
				// discover repos from source Harbor API
				list, err := srcHC.ListRepos(syncProject)
				if err != nil {
					return fmt.Errorf("list repos: %w", err)
				}
				for _, r := range list {
					// r.Name is "project/repo"
					parts := strings.SplitN(r.Name, "/", 2)
					repo := parts[len(parts)-1]

					// apply include/exclude
					if len(p.Includes) > 0 && !globAny(p.Includes, repo) {
						continue
					}
					if len(p.Excludes) > 0 && globAny(p.Excludes, repo) {
						continue
					}

					tgs := p.Tags
					if len(tgs) == 0 {
						tgs = []string{"*"}
					}
					plan = append(plan, planItem{Repo: repo, Tags: tgs})
				}
			}
			if len(plan) == 0 {
				color.Yellow("No repos matched by %s for project %q.", syncRulesPath, syncProject)
				if dryRun {
					return nil
				}
			}
		} else {
			// legacy --repo / --tags path
			var repos []string
			if syncRepo == "" {
				list, err := srcHC.ListRepos(syncProject)
				if err != nil {
					return fmt.Errorf("list repos: %w", err)
				}
				for _, r := range list {
					parts := strings.SplitN(r.Name, "/", 2)
					repo := parts[len(parts)-1]
					repos = append(repos, repo)
				}
			} else {
				repos = []string{syncRepo}
			}
			tgs := syncTags
			if len(tgs) == 0 {
				tgs = []string{"*"}
			}
			for _, r := range repos {
				plan = append(plan, planItem{Repo: r, Tags: tgs})
			}
		}

		// Resolve registry endpoints for copy
		srcReg := fr.RegistryURL
		if srcReg == "" {
			srcReg = fr.URL // fallback for real Harbor where same host
		}
		dstReg := tr.RegistryURL
		if dstReg == "" {
			dstReg = tr.URL
		}

		// Build copy tasks
		var tasks []copyTask

		for _, item := range plan {
			arts, err := srcHC.ListArtifacts(syncProject, item.Repo)
			if err != nil {
				color.Red("skip %s/%s: %v", syncProject, item.Repo, err)
				continue
			}

			for _, a := range arts {
				for _, tg := range a.Tags {
					tag := tg.Name
					if !globAny(item.Tags, tag) {
						continue
					}

					srcRef := fmt.Sprintf("docker://%s/%s/%s:%s",
						trimScheme(srcReg), syncProject, item.Repo, tag)
					dstRef := fmt.Sprintf("docker://%s/%s/%s:%s",
						trimScheme(dstReg), syncProject, item.Repo, tag)

					if dryRun {
						color.Yellow("[dry-run] skopeo copy %s -> %s", srcRef, dstRef)
						continue
					}
					tasks = append(tasks, copyTask{srcRef: srcRef, dstRef: dstRef})
				}
			}
		}

		// Execute tasks with worker pool
		if !dryRun {
			runCopies(tasks, cfg, syncDockerNetwork, fr.Insecure, tr.Insecure, fu, fp, tu, tp)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().StringVar(&syncProject, "project", "", "Project to sync")
	syncCmd.Flags().StringVar(&syncRepo, "repo", "", "Specific repo to sync (optional)")
	syncCmd.Flags().StringSliceVar(&syncTags, "tags", nil, "Tag globs to include (default: all)")
	syncCmd.Flags().BoolVar(&dryRun, "dry-run", true, "Print what would be copied, do not execute")
	syncCmd.Flags().StringVar(&syncDockerNetwork, "docker-network", "", "Docker network for skopeo")
	syncCmd.Flags().StringVar(&syncRulesPath, "rules", "", "Path to rules.yaml (overrides --repo/--tags)")
	syncCmd.Flags().IntVar(&maxConcurrent, "concurrency", 2, "Number of parallel copy operations")
}

// ----- worker pool -----
func runCopies(tasks []copyTask, cfg *config.Config, dockerNetwork string,
	srcInsecure, dstInsecure bool, fu, fp, tu, tp string) {

	if len(tasks) == 0 {
		color.Green("Nothing to copy.")
		return
	}
	if maxConcurrent < 1 {
		maxConcurrent = 1
	}

	bar := progressbar.NewOptions(len(tasks),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetWidth(20),
		progressbar.OptionShowCount(),
		progressbar.OptionSetDescription("[cyan]Copying images...[reset]"),
		progressbar.OptionShowElapsedTimeOnFinish(),
		progressbar.OptionClearOnFinish(),
	)

	taskCh := make(chan copyTask)
	doneCh := make(chan struct{})

	// spawn workers
	for i := 0; i < maxConcurrent; i++ {
		go func() {
			for t := range taskCh {
				args := buildSkopeoCopyArgs(
					cfg.SkopeoPath,
					dockerNetwork,
					srcInsecure, dstInsecure,
					fu, fp, tu, tp,
					t.srcRef, t.dstRef,
				)

				out, err := shell.Run(cfg.SkopeoPath, args...)
				if err != nil {
					if strings.Contains(out+err.Error(), "manifest unknown") {
						color.Yellow("skip (missing on source): %s", t.srcRef)
					} else {
						color.Red("copy failed: %v", err)
					}
				}
				bar.Add(1)
			}
			doneCh <- struct{}{}
		}()
	}

	// feed tasks
	for _, t := range tasks {
		taskCh <- t
	}
	close(taskCh)

	// wait all
	for i := 0; i < maxConcurrent; i++ {
		<-doneCh
	}

	bar.Finish()
	color.Green("Completed %d copy operation(s).", len(tasks))
	time.Sleep(200 * time.Millisecond) // smooth finish
}

// ----- helpers -----

func buildSkopeoCopyArgs(
	skopeoPath string,
	dockerNetwork string,
	srcInsecure, dstInsecure bool,
	srcUser, srcPass, dstUser, dstPass string,
	srcRef, dstRef string,
) []string {
	if strings.EqualFold(skopeoPath, "docker") {
		args := []string{"run", "--rm"}
		if dockerNetwork != "" {
			args = append(args, "--network", dockerNetwork)
		}
		image := "quay.io/skopeo/stable"
		args = append(args, image, "copy")
		if srcInsecure {
			args = append(args, "--src-tls-verify=false")
		}
		if dstInsecure {
			args = append(args, "--dest-tls-verify=false")
		}
		if srcUser != "" || srcPass != "" {
			args = append(args, "--src-creds", fmt.Sprintf("%s:%s", srcUser, srcPass))
		}
		if dstUser != "" || dstPass != "" {
			args = append(args, "--dest-creds", fmt.Sprintf("%s:%s", dstUser, dstPass))
		}
		args = append(args, srcRef, dstRef)
		return args
	}

	// Native skopeo on PATH
	args := []string{"copy"}
	if srcInsecure {
		args = append(args, "--src-tls-verify=false")
	}
	if dstInsecure {
		args = append(args, "--dest-tls-verify=false")
	}
	if srcUser != "" || srcPass != "" {
		args = append(args, "--src-creds", fmt.Sprintf("%s:%s", srcUser, srcPass))
	}
	if dstUser != "" || dstPass != "" {
		args = append(args, "--dest-creds", fmt.Sprintf("%s:%s", dstUser, dstPass))
	}
	args = append(args, srcRef, dstRef)
	return args
}

func trimScheme(u string) string {
	if strings.HasPrefix(u, "http://") {
		return strings.TrimPrefix(u, "http://")
	}
	if strings.HasPrefix(u, "https://") {
		return strings.TrimPrefix(u, "https://")
	}
	return u
}

func globAny(patterns []string, s string) bool {
	if len(patterns) == 0 {
		return true
	}
	for _, p := range patterns {
		re := "^" + regexp.QuoteMeta(strings.TrimSpace(p)) + "$"
		re = strings.ReplaceAll(re, `\*`, ".*")
		if regexp.MustCompile(re).MatchString(s) {
			return true
		}
	}
	return false
}
