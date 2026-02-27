package main

import (
	"fmt"
	"os"
	"strings"

	"kutil/internal/kflap"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "kflap",
		Short: "Kubernetes resource flapping detector",
		Long:  "Monitor Kubernetes resources for excessive updates by tracking resourceVersion changes",
	}

	resourcesCmd := &cobra.Command{
		Use:   "resources",
		Short: "Monitor resource versions in real-time",
		Long:  "Display a live table of Kubernetes resources and their resourceVersion changes",
		Run:   runResources,
	}

	// Add flags
	resourcesCmd.Flags().StringP("resources", "r", "", "Comma-delimited list of resource types to monitor (default: all)")
	resourcesCmd.Flags().IntP("interval", "i", 5, "Polling interval in seconds")
	resourcesCmd.Flags().StringP("namespaces", "n", "", "Comma-delimited list of namespaces to monitor (default: all)")
	resourcesCmd.Flags().IntP("limit", "l", 20, "Number of table rows to display")

	rootCmd.AddCommand(resourcesCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runResources(cmd *cobra.Command, args []string) {
	// Parse flags
	resourcesFlag, _ := cmd.Flags().GetString("resources")
	interval, _ := cmd.Flags().GetInt("interval")
	namespacesFlag, _ := cmd.Flags().GetString("namespaces")
	limit, _ := cmd.Flags().GetInt("limit")

	// Parse comma-delimited values
	var resources []string
	if resourcesFlag != "" {
		resources = strings.Split(resourcesFlag, ",")
		for i := range resources {
			resources[i] = strings.TrimSpace(resources[i])
		}
	}

	var namespaces []string
	if namespacesFlag != "" {
		namespaces = strings.Split(namespacesFlag, ",")
		for i := range namespaces {
			namespaces[i] = strings.TrimSpace(namespaces[i])
		}
	}

	// Create config
	config := kflap.Config{
		Resources:  resources,
		Namespaces: namespaces,
		Interval:   interval,
		Limit:      limit,
	}

	// Run the TUI
	if err := kflap.Run(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
