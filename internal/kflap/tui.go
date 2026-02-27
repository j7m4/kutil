package kflap

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	headerStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true)

	changesStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("9")) // Red

	cellStyle = lipgloss.NewStyle().
		PaddingRight(2)
)

type tickMsg time.Time

type model struct {
	monitor   *Monitor
	config    Config
	resources []*ResourceInfo
	err       error
	ready     bool
}

func newModel(monitor *Monitor, config Config) model {
	return model{
		monitor: monitor,
		config:  config,
	}
}

func (m model) Init() tea.Cmd {
	// Start polling immediately
	return tea.Batch(
		tickCmd(m.config.Interval),
		doPoll(m.monitor),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}

	case tickMsg:
		// Poll on each tick
		return m, tea.Batch(
			tickCmd(m.config.Interval),
			doPoll(m.monitor),
		)

	case pollResultMsg:
		m.resources = msg.resources
		m.err = msg.err
		m.ready = true
		return m, nil

	case tea.WindowSizeMsg:
		return m, nil
	}

	return m, nil
}

func (m model) View() string {
	if !m.ready {
		return "Loading...\n"
	}

	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress 'q' to quit.\n", m.err)
	}

	var b strings.Builder

	// Title
	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Kubernetes Resource Monitor"))
	b.WriteString("\n\n")

	// Sort resources by changes(descending), resourceVersion (descending)
	sortedResources := make([]*ResourceInfo, len(m.resources))
	copy(sortedResources, m.resources)
	sort.Slice(sortedResources, func(i, j int) bool {
		if sortedResources[i].Changes != sortedResources[j].Changes {
			return sortedResources[i].Changes > sortedResources[j].Changes
		}
		return sortedResources[i].ResourceVersion > sortedResources[j].ResourceVersion
	})

	// Limit to configured number of rows
	if len(sortedResources) > m.config.Limit {
		sortedResources = sortedResources[:m.config.Limit]
	}

	// Calculate column widths
	nameWidth := 30
	typeWidth := 25
	nsWidth := 20
	versionWidth := 20
	changesWidth := 10

	// Render table header
	header := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s",
		nameWidth, "NAME",
		typeWidth, "TYPE",
		nsWidth, "NAMESPACE",
		versionWidth, "RESOURCE VERSION",
		changesWidth, "CHANGES",
	)
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n")

	// Render table rows
	for _, info := range sortedResources {
		name := truncate(info.Name, nameWidth)
		resourceType := truncate(info.Type, typeWidth)
		namespace := truncate(info.Namespace, nsWidth)
		if namespace == "" {
			namespace = "<cluster>"
		}
		version := fmt.Sprintf("%d", info.ResourceVersion)
		changes := fmt.Sprintf("%d", info.Changes)

		// Format the row
		row := fmt.Sprintf("%-*s %-*s %-*s %-*s ",
			nameWidth, name,
			typeWidth, resourceType,
			nsWidth, namespace,
			versionWidth, version,
		)

		// Apply styling to delta if positive
		if info.Changes > 0 {
			deltaFormatted := changesStyle.Render(fmt.Sprintf("%-*s", changesWidth, changes))
			b.WriteString(row)
			b.WriteString(deltaFormatted)
		} else {
			b.WriteString(row)
			b.WriteString(fmt.Sprintf("%-*s", changesWidth, changes))
		}

		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Showing top %d resources sorted by changes(DESC), resourceVersion(DESC)\n", len(sortedResources)))
	b.WriteString(fmt.Sprintf("Polling interval: %d seconds\n", m.config.Interval))
	b.WriteString("\nPress 'q' to quit.\n")

	return b.String()
}

func tickCmd(intervalSecs int) tea.Cmd {
	return tea.Tick(time.Duration(intervalSecs)*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type pollResultMsg struct {
	resources []*ResourceInfo
	err       error
}

func doPoll(monitor *Monitor) tea.Cmd {
	return func() tea.Msg {
		err := monitor.Poll()
		resources := monitor.GetResources()
		return pollResultMsg{
			resources: resources,
			err:       err,
		}
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
