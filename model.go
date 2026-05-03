package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	activeTab    Tab
	packages     []Package
	selected     int
	showDetail   bool
	detailInfo   string
	loading      bool
	statusMsg    string
	inputMode    InputMode
	searchInput  textinput.Model
	filterInput  textinput.Model
	spinner      spinner.Model
	aurHelper    string
	installed    []Package
	updates      []Package
	width        int
	height       int
	ready        bool
	helpVisible  bool
}

func NewModel() Model {
	si := textinput.New()
	si.Placeholder = "Search packages..."
	si.CharLimit = 100
	si.Width = 40

	fi := textinput.New()
	fi.Placeholder = "Filter installed packages..."
	fi.CharLimit = 100
	fi.Width = 40

	s := spinner.New()
	s.Style = spinnerStyle
	s.Spinner = spinner.Dot

	return Model{
		activeTab:   TabSearch,
		searchInput: si,
		filterInput: fi,
		spinner:     s,
		inputMode:   ModeNavigate,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			ah := detectAURHelper()
			pkgs, err := listInstalled()
			if err != nil {
				return Msg{Kind: MsgError, Payload: err}
			}
			return Msg{Kind: MsgInstalledList, Payload: struct {
				pkgs      []Package
				aurHelper string
			}{pkgs, ah}}
		},
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

	case tea.KeyMsg:
		return m.handleKey(msg)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case Msg:
		switch msg.Kind {
		case MsgInstalledList:
			d := msg.Payload.(struct {
				pkgs      []Package
				aurHelper string
			})
			m.installed = d.pkgs
			m.aurHelper = d.aurHelper
			if m.activeTab == TabInstalled {
				m.packages = d.pkgs
			}
			m.loading = false
			m.statusMsg = fmt.Sprintf("Loaded %d packages", len(m.installed))

		case MsgSearchResult:
			m.packages = msg.Payload.([]Package)
			m.selected = 0
			m.loading = false
			m.statusMsg = fmt.Sprintf("Found %d packages", len(m.packages))

		case MsgUpdatesList:
			m.updates = msg.Payload.([]Package)
			if m.activeTab == TabUpdates {
				m.packages = m.updates
			}
			m.loading = false
			m.statusMsg = fmt.Sprintf("Updates: %d", len(m.updates))

		case MsgPackageInfo:
			m.detailInfo = msg.Payload.(string)
			m.showDetail = true
			m.loading = false

		case MsgCmdResult:
			m.statusMsg = msg.Payload.(string)

		case MsgError:
			m.loading = false
			m.statusMsg = fmt.Sprintf("Error: %v", msg.Payload.(error))
		}

	case error:
		m.loading = false
		m.statusMsg = fmt.Sprintf("Error: %v", msg)
	}

	return m, nil
}

// handleKey processes all key events and returns (updated model, cmd).
// This is a value receiver on Model — no pointer shenanigans.
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// ── Help overlay ──
	if m.helpVisible {
		if msg.String() == "?" || msg.String() == "esc" {
			m.helpVisible = false
		}
		return m, nil
	}

	// ── Toggle help ──
	if msg.String() == "?" {
		m.helpVisible = true
		return m, nil
	}

	// ── Quit ──
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	// ── Detail view ──
	if m.showDetail {
		if msg.String() == "enter" || msg.String() == "esc" {
			m.showDetail = false
		}
		return m, nil
	}

	// ── Tab switching (always available) ──
	switch msg.String() {
	case "1":
		return m.switchToTab(TabSearch)
	case "2":
		return m.switchToTab(TabInstalled)
	case "3":
		return m.switchToTab(TabUpdates)
	}

	// ── Search input mode ──
	if m.inputMode == ModeSearch {
		return m.handleSearchInput(msg)
	}

	// ── Filter input mode ──
	if m.inputMode == ModeFilter {
		return m.handleFilterInput(msg)
	}

	// ── Navigation mode ──
	return m.handleNav(msg)
}

func (m Model) handleSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		q := strings.TrimSpace(m.searchInput.Value())
		m.inputMode = ModeNavigate
		if q == "" {
			m.searchInput.Blur()
			return m, nil
		}
		m.searchInput.Blur()
		m.statusMsg = fmt.Sprintf("Searching: %s...", q)
		return m.execSearch(q)

	case tea.KeyEscape:
		m.searchInput.SetValue("")
			m.inputMode = ModeNavigate
			m.searchInput.Blur()
		return m, nil
	}

	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	return m, cmd
}

func (m Model) handleFilterInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyEscape {
		m.filterInput.SetValue("")
		m.inputMode = ModeNavigate
		m.filterInput.Blur()
		m.packages = m.installed
		return m, nil
	}

	var cmd tea.Cmd
	m.filterInput, cmd = m.filterInput.Update(msg)

	q := strings.ToLower(m.filterInput.Value())
	if q == "" {
		m.packages = m.installed
	} else {
		var f []Package
		for _, p := range m.installed {
			if strings.Contains(strings.ToLower(p.Name), q) ||
				strings.Contains(strings.ToLower(p.Description), q) {
				f = append(f, p)
			}
		}
		m.packages = f
	}
	return m, cmd
}

func (m Model) handleNav(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// ── Cursor movement (type-based) ──
	switch msg.Type {
	case tea.KeyUp:
		if m.selected > 0 {
			m.selected--
		}
	case tea.KeyDown:
		if m.selected < len(m.packages)-1 {
			m.selected++
		}
	case tea.KeyPgUp:
		m.selected -= 10
		if m.selected < 0 {
			m.selected = 0
		}
	case tea.KeyPgDown:
		m.selected += 10
		if m.selected >= len(m.packages) {
			m.selected = len(m.packages) - 1
		}
	case tea.KeyHome:
		m.selected = 0
	case tea.KeyEnd:
		m.selected = len(m.packages) - 1
	}

	// ── Action keys (string-based) ──
	switch msg.String() {
	case "j":
		if m.selected < len(m.packages)-1 {
			m.selected++
		}
	case "k":
		if m.selected > 0 {
			m.selected--
		}
	case "q":
		return m, tea.Quit

		case "/":
			switch m.activeTab {
			case TabSearch:
				m.searchInput.SetValue("")
				m.inputMode = ModeSearch
				return m, m.searchInput.Focus()
			case TabInstalled:
				m.filterInput.SetValue("")
				m.inputMode = ModeFilter
				return m, m.filterInput.Focus()
			default:
				m.statusMsg = "Search not available in this tab"
			}

	case "g":
		m.selected = 0
	case "G":
		m.selected = len(m.packages) - 1

	case "enter":
		if len(m.packages) == 0 {
			return m, nil
		}
		pkg := m.packages[m.selected]
		return m.execInfo(pkg)

	case "i":
		if len(m.packages) == 0 {
			return m, nil
		}
		pkg := m.packages[m.selected]
		if pkg.InstalledVersion != "" {
			m.statusMsg = fmt.Sprintf("%s already installed", pkg.Name)
			return m, nil
		}
		return m.execInstall(pkg)

	case "r":
		if len(m.packages) == 0 {
			return m, nil
		}
		pkg := m.packages[m.selected]
		if pkg.InstalledVersion == "" && m.activeTab != TabUpdates {
			m.statusMsg = fmt.Sprintf("%s not installed", pkg.Name)
			return m, nil
		}
		return m.execRemove(pkg)

	case "u":
		if len(m.packages) == 0 {
			return m, nil
		}
		return m.execUpdate(m.packages[m.selected])

	case "U":
		return m.execUpgrade()

	case "R":
		return m.execReload()
	}

	return m, nil
}

// ── Tab switching ──

func (m Model) switchToTab(tab Tab) (tea.Model, tea.Cmd) {
	m.selected = 0
	m.showDetail = false
	m.searchInput.Blur()
	m.filterInput.Blur()

	switch tab {
	case TabSearch:
		m.activeTab = TabSearch
		m.packages = nil
		m.statusMsg = "Press / to search"

	case TabInstalled:
		m.activeTab = TabInstalled
		m.packages = m.installed
		m.statusMsg = fmt.Sprintf("Installed: %d", len(m.installed))

	case TabUpdates:
		m.activeTab = TabUpdates
		return m, m.cmdCheckUpdates()
	}

	return m, nil
}

// ── Async commands ──

func (m Model) cmdCheckUpdates() tea.Cmd {
	m.loading = true
	m.statusMsg = "Checking updates..."
	return func() tea.Msg {
		pkgs, err := checkUpdates(m.aurHelper)
		if err != nil {
			return Msg{Kind: MsgError, Payload: err}
		}
		return Msg{Kind: MsgUpdatesList, Payload: pkgs}
	}
}

func (m Model) execSearch(query string) (tea.Model, tea.Cmd) {
	m.loading = true
	ah := m.aurHelper
	return m, func() tea.Msg {
		pkgs, err := searchPackages(query, ah)
		if err != nil {
			return Msg{Kind: MsgError, Payload: err}
		}
		return Msg{Kind: MsgSearchResult, Payload: pkgs}
	}
}

func (m Model) execInfo(pkg Package) (tea.Model, tea.Cmd) {
	m.loading = true
	ah := m.aurHelper
	return m, func() tea.Msg {
		info, err := getPackageInfo(pkg, ah)
		if err != nil {
			return Msg{Kind: MsgError, Payload: err}
		}
		return Msg{Kind: MsgPackageInfo, Payload: info}
	}
}

func (m Model) execInstall(pkg Package) (tea.Model, tea.Cmd) {
	m.statusMsg = fmt.Sprintf("Installing %s...", pkg.Name)
	ah := m.aurHelper
	return m, func() tea.Msg {
		_, err := installPkg(pkg, ah)
		if err != nil {
			return Msg{Kind: MsgError, Payload: fmt.Errorf("install failed: %v", err)}
		}
		return Msg{Kind: MsgCmdResult, Payload: fmt.Sprintf("Installed %s", pkg.Name)}
	}
}

func (m Model) execRemove(pkg Package) (tea.Model, tea.Cmd) {
	m.statusMsg = fmt.Sprintf("Removing %s...", pkg.Name)
	return m, func() tea.Msg {
		_, err := removePkg(pkg)
		if err != nil {
			return Msg{Kind: MsgError, Payload: fmt.Errorf("remove failed: %v", err)}
		}
		return Msg{Kind: MsgCmdResult, Payload: fmt.Sprintf("Removed %s", pkg.Name)}
	}
}

func (m Model) execUpdate(pkg Package) (tea.Model, tea.Cmd) {
	m.statusMsg = fmt.Sprintf("Updating %s...", pkg.Name)
	ah := m.aurHelper
	return m, func() tea.Msg {
		_, err := updatePkg(pkg, ah)
		if err != nil {
			return Msg{Kind: MsgError, Payload: fmt.Errorf("update failed: %v", err)}
		}
		return Msg{Kind: MsgCmdResult, Payload: fmt.Sprintf("Updated %s", pkg.Name)}
	}
}

func (m Model) execUpgrade() (tea.Model, tea.Cmd) {
	m.statusMsg = "System upgrade..."
	ah := m.aurHelper
	return m, func() tea.Msg {
		out, err := systemUpgrade(ah)
		if err != nil {
			return Msg{Kind: MsgError, Payload: fmt.Errorf("upgrade failed: %v", err)}
		}
		return Msg{Kind: MsgCmdResult, Payload: fmt.Sprintf("Upgrade done:\n%s", out)}
	}
}

func (m Model) execReload() (tea.Model, tea.Cmd) {
	if m.activeTab == TabUpdates {
		m.statusMsg = "Already on Updates tab"
		return m, nil
	}
	m.loading = true
	ah := m.aurHelper
	return m, func() tea.Msg {
		pkgs, err := listInstalled()
		if err != nil {
			return Msg{Kind: MsgError, Payload: err}
		}
		return Msg{Kind: MsgInstalledList, Payload: struct {
			pkgs      []Package
			aurHelper string
		}{pkgs, ah}}
	}
}
