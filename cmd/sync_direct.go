package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/hakantongur/harair/internal/config"
	"github.com/hakantongur/harair/internal/shell"
	"github.com/spf13/cobra"
)

var (
	fromRef       string
	toRef         string
	srcInsecure   bool
	destInsecure  bool
	reallyDoCopy  bool
	dockerNetwork string
)

var syncDirectCmd = &cobra.Command{
	Use:   "sync-direct",
	Short: "Directly copy one image ref to another (no Harbor discovery)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if fromRef == "" || toRef == "" {
			return fmt.Errorf("please pass --from and --to (e.g., docker://localhost:5001/demo/demo-repo:latest)")
		}
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return err
		}
		color.Cyan("Plan: skopeo copy %s -> %s", fromRef, toRef)
		if !reallyDoCopy {
			color.Yellow("[dry-run] Not executing. Add --do to perform the copy.")
			return nil
		}

		if cfg.SkopeoPath == "docker" || cfg.SkopeoPath == "Docker" {
			args = append(args, "run", "--rm")
			// 👇 add the network if requested
			if dockerNetwork != "" {
				args = append(args, "--network", dockerNetwork)
			}
			args = append(args, "quay.io/skopeo/stable", "copy")
		} else {
			args = append(args, "copy")
		}
		if srcInsecure {
			args = append(args, "--src-tls-verify=false")
		}
		if destInsecure {
			args = append(args, "--dest-tls-verify=false")
		}
		args = append(args, fromRef, toRef)

		color.Green("Executing: %s %v", cfg.SkopeoPath, args)
		out, err := shell.Run(cfg.SkopeoPath, args...)
		if err != nil {
			return fmt.Errorf("copy failed: %w", err)
		}
		if len(out) > 0 {
			fmt.Println(out)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(syncDirectCmd)
	syncDirectCmd.Flags().StringVar(&fromRef, "from", "", "Source image ref (e.g., docker://localhost:5001/demo/repo:tag)")
	syncDirectCmd.Flags().StringVar(&toRef, "to", "", "Destination image ref (e.g., docker://localhost:5002/demo/repo:tag)")
	syncDirectCmd.Flags().BoolVar(&srcInsecure, "src-insecure", true, "Disable TLS verify for source")
	syncDirectCmd.Flags().BoolVar(&destInsecure, "dst-insecure", true, "Disable TLS verify for destination")
	syncDirectCmd.Flags().BoolVar(&reallyDoCopy, "do", false, "Actually perform the copy (otherwise dry-run)")
	syncDirectCmd.Flags().StringVar(&dockerNetwork, "docker-network", "", "Docker network to run skopeo on (container mode)")
}
