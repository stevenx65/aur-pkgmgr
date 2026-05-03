package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED")).
			Background(lipgloss.Color("#1A1A2E")).
			Padding(0, 1)

	tabStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(lipgloss.Color("#888888"))

	activeTabStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7C3AED")).
			Bold(true)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#0055AA")).
			Bold(true).Padding(0, 1)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CCCCCC"))

	installedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#50C878"))

	aurStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF8C00"))

	updateStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B"))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AAAAAA")).
			Background(lipgloss.Color("#1A1A2E")).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	spinnerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C3AED"))

	detailStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7C3AED")).
			Padding(1, 2).
			Background(lipgloss.Color("#1E1E32"))

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#2D2D44")).
			Padding(0, 1)

	titleBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#1A1A2E")).
			Padding(0, 1)
)

func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}
	return lipgloss.JoinVertical(lipgloss.Top,
		m.renderHeader(),
		m.renderContent(),
		m.renderStatus(),
	)
}

func (m Model) renderHeader() string {
	return titleBarStyle.Render(func() string {
		var sb strings.Builder
		sb.WriteString(titleStyle.Render(" PkgMgr "))
		tabs := []struct {
			name string
			tab  Tab
		}{
			{"Search", TabSearch},
			{"Installed", TabInstalled},
			{"Updates", TabUpdates},
		}
		for _, t := range tabs {
			if m.activeTab == t.tab {
				sb.WriteString(" ")
				sb.WriteString(activeTabStyle.Render(t.name))
			} else {
				sb.WriteString(" ")
				sb.WriteString(tabStyle.Render(t.name))
			}
		}
		return sb.String()
	}())
}

func (m Model) renderContent() string {
	if m.helpVisible {
		return m.renderHelp()
	}
	if m.loading {
		return m.renderLoading()
	}
	if m.showDetail {
		return m.renderDetail()
	}
	return m.renderList()
}

func (m Model) renderLoading() string {
	msg := m.statusMsg
	if msg == "" {
		msg = "Loading..."
	}
	return lipgloss.Place(m.width, m.height-3, lipgloss.Center, lipgloss.Center,
		m.spinner.View()+" "+msg,
	)
}

func (m Model) renderList() string {
	var b strings.Builder

	// Search/filter bar (1 line)
	switch m.activeTab {
	case TabSearch:
		if m.inputMode == ModeSearch {
			b.WriteString(inputStyle.Width(m.width-2).Render(" Search: "+m.searchInput.View()))
		} else {
			b.WriteString(helpStyle.Width(m.width-2).Render(fmt.Sprintf("Press / to search  |  %d results", len(m.packages))))
		}
	case TabInstalled:
		if m.inputMode == ModeFilter {
			b.WriteString(inputStyle.Width(m.width-2).Render(" Filter: "+m.filterInput.View()))
		} else {
			b.WriteString(helpStyle.Width(m.width-2).Render(fmt.Sprintf("Installed: %d  |  Press / to filter", len(m.installed))))
		}
	case TabUpdates:
		b.WriteString(helpStyle.Width(m.width-2).Render(fmt.Sprintf("Updates: %d  |  Press U to upgrade all", len(m.packages))))
	}
	b.WriteString("\n")

	if len(m.packages) == 0 {
		msg := "No results"
		if m.activeTab == TabSearch {
			msg = "Press / to search for packages"
		} else if m.activeTab == TabUpdates {
			msg = "System is up to date!"
		}
		b.WriteString(helpStyle.Render("  " + msg))
		return b.String()
	}

	// list occupies remaining space: height - 3 (header) - 1 (search bar) - 1 (status)
	listH := m.height - 5
	if listH < 1 {
		listH = 1
	}

	start, end := visibleRange(m.selected, len(m.packages), listH)

	for i := start; i < end; i++ {
		if i >= len(m.packages) {
			break
		}
		line := renderPkgLine(m.packages[i], i == m.selected, m.width)
		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}

func visibleRange(sel, total, h int) (int, int) {
	if h <= 0 {
		h = 10
	}
	if sel < 0 {
		sel = 0
	}
	if sel >= total {
		sel = total - 1
	}
	if total == 0 {
		return 0, 0
	}
	// center the selection
	start := sel - h/2
	if start < 0 {
		start = 0
	}
	end := start + h
	if end > total {
		end = total
		start = end - h
		if start < 0 {
			start = 0
		}
	}
	return start, end
}

func renderPkgLine(pkg Package, selected bool, width int) string {
	style := normalStyle
	prefix := "  "
	if selected {
		style = selectedStyle
		prefix = "▸ "
	}

	// construct line parts
	line := fmt.Sprintf("%s%-28s %-14s %-50s",
		prefix,
		truncate(pkg.Name, 28),
		truncate(pkg.Version, 14),
		truncate(pkg.Description, 50),
	)

	// tags
	var tags []string
	if pkg.InstalledVersion != "" {
		tags = append(tags, installedStyle.Render("installed"))
	}
	switch pkg.Repository {
	case "aur":
		tags = append(tags, aurStyle.Render("aur"))
	case "core":
		tags = append(tags, lipgloss.NewStyle().Foreground(lipgloss.Color("#60A5FA")).Render("core"))
	case "extra":
		tags = append(tags, lipgloss.NewStyle().Foreground(lipgloss.Color("#A78BFA")).Render("extra"))
	case "":
		// local package, no repo tag
	default:
		tags = append(tags, helpStyle.Render(pkg.Repository))
	}
	for _, t := range tags {
		line += " " + t
	}

	return style.Width(width - 2).Render(line)
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max-1] + "…"
	}
	return s
}

func (m Model) renderDetail() string {
	info := m.detailInfo
	if info == "" {
		info = "No info"
	}
	box := detailStyle.Width(m.width-4).Height(m.height-6).Render(info)
	return lipgloss.JoinVertical(lipgloss.Top, box, helpStyle.Render(" Press Enter/Esc to go back"))
}

func (m Model) renderStatus() string {
	status := m.statusMsg
	if status == "" {
		status = "Ready"
	}
	sel := ""
	if len(m.packages) > 0 {
		sel = fmt.Sprintf("  [%d/%d]", m.selected+1, len(m.packages))
	}
	return statusStyle.Width(m.width-2).Render(status + sel)
}

func (m Model) renderHelp() string {
	helpTxt := `KEYBINDINGS
━━━━━━━━━━━
1/2/3      Switch tabs (Search | Installed | Updates)
j/k ↑/↓   Navigate list       g/G       Top / Bottom
/          Search packages     Enter     Package details
i          Install package     r         Remove package
u          Update package      U         Full system upgrade
R          Reload              ?         Help
q          Quit                Esc       Back / Cancel`
	return lipgloss.Place(m.width, m.height-3, lipgloss.Center, lipgloss.Center,
		detailStyle.Width(m.width-8).Render(helpTxt))
}
