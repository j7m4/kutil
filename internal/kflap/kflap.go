package kflap

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Config holds the configuration for the kflap monitor
type Config struct {
	Resources  []string // Resource types to monitor (empty = all)
	Namespaces []string // Namespaces to monitor (empty = all)
	Interval   int      // Polling interval in seconds
	Limit      int      // Number of rows to display
}

// Run starts the kflap TUI
func Run(config Config) error {
	// Create monitor
	monitor, err := NewMonitor(config)
	if err != nil {
		return err
	}

	// Create and start the TUI program
	p := tea.NewProgram(newModel(monitor, config))
	_, err = p.Run()
	return err
}
