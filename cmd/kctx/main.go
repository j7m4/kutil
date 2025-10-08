package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"kutil/internal/kctx"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "kctx",
		Short: "Kubernetes context utility",
	}

	lsCmd := &cobra.Command{
		Use:   "ls",
		Short: "List all kubernetes contexts",
		Run:   lsContexts,
	}

	trCmd := &cobra.Command{
		Use:   "tr [INPUT_REGEX] [REPLACEMENT_VALUE]",
		Short: "Transform context names using regex patterns",
		Long: `Transform context names using regex patterns.

Usage:
  kctx tr INPUT_REGEX REPLACEMENT_VALUE  Replace matched regex with replacement value
  kctx tr -d DELETION_REGEX              Delete matched regex from context names`,
		Args: cobra.MinimumNArgs(1),
		Run:  trContexts,
	}

	trCmd.Flags().BoolP("delete", "d", false, "Delete matched regex instead of replacing")
	trCmd.Flags().BoolP("force", "f", false, "Apply changes without confirmation prompt")

	grepCmd := &cobra.Command{
		Use:   "grep REGEX",
		Short: "Filter contexts by regex pattern",
		Long:  "Filter and display contexts that match the given regex pattern",
		Args:  cobra.ExactArgs(1),
		Run:   grepContexts,
	}

	grepCmd.Flags().BoolP("invert-match", "v", false, "Show contexts that do NOT match the pattern")

	backupCmd := &cobra.Command{
		Use:   "backup",
		Short: "Create a timestamped backup of the kubeconfig file",
		Long:  "Create a backup copy of the kubeconfig file with a timestamp suffix",
		Run:   backupKubeconfig,
	}

	rootCmd.AddCommand(lsCmd, trCmd, grepCmd, backupCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func lsContexts(_ *cobra.Command, _ []string) {
	if err := kctx.ListContexts(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func trContexts(cmd *cobra.Command, args []string) {
	deleteMode, _ := cmd.Flags().GetBool("delete")
	force, _ := cmd.Flags().GetBool("force")

	var err error
	if deleteMode {
		if len(args) != 1 {
			fmt.Fprintf(os.Stderr, "Error: delete mode requires exactly one regex argument\n")
			os.Exit(1)
		}
		err = kctx.DeleteFromContexts(args[0], force)
	} else {
		if len(args) != 2 {
			fmt.Fprintf(os.Stderr, "Error: replace mode requires exactly two arguments: INPUT_REGEX REPLACEMENT_VALUE\n")
			os.Exit(1)
		}
		err = kctx.ReplaceInContexts(args[0], args[1], force)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func grepContexts(cmd *cobra.Command, args []string) {
	regex := args[0]
	invertMatch, _ := cmd.Flags().GetBool("invert-match")

	if err := kctx.GrepContexts(regex, invertMatch); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func backupKubeconfig(_ *cobra.Command, _ []string) {
	if err := kctx.BackupKubeconfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
