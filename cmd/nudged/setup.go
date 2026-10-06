package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Spectral-Kaburu/annoying-sister/internal/config"
	"github.com/Spectral-Kaburu/annoying-sister/internal/tts"
)

// runSetup launches the interactive setup wizard that lets the user pick
// directories for scan_roots (and tweak ignored_paths) and writes the
// result to config.json. It is invoked via "nudged -setup".
func runSetup(cfgPath string) error {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	m, err := newSetupModel(cfgPath, cfg)
	if err != nil {
		return err
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return err
	}

	sm := final.(setupModel)
	if sm.err != nil {
		return fmt.Errorf("save config: %w", sm.err)
	}
	if sm.saved {
		fmt.Fprintf(os.Stderr, "\n✓ Saved %d scan root(s) to %s\n", len(sm.roots), cfgPath)

		// Discover and test SpectreTTS socket by sending a test speech
		sock := tts.DiscoverSocket(sm.cfg.SpectreTTSSocketPath)
		fmt.Fprintf(os.Stderr, "🔍 Testing SpectreTTS socket (%s)...\n", sock)
		client := tts.NewClient(sock)
		if err := client.Speak("hello"); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Warning: could not speak to SpectreTTS socket at %s: %v\n", sock, err)
		} else {
			fmt.Fprintf(os.Stderr, "✓ SpectreTTS test successful: spoke \"hello\"\n")
		}
	} else {
		fmt.Fprintln(os.Stderr, "\nSetup cancelled — no changes written.")
	}
	return nil
}

// ── key bindings ─────────────────────────────────────────────────────────────

type setupKeys struct {
	Add    key.Binding
	Remove key.Binding
	Save   key.Binding
	Quit   key.Binding
	Switch key.Binding
	Up     key.Binding
	Down   key.Binding
}

var keys = setupKeys{
	Add: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "add directory"),
	),
	Remove: key.NewBinding(
		key.WithKeys("d", "backspace"),
		key.WithHelp("d/⌫", "remove selected"),
	),
	Save: key.NewBinding(
		key.WithKeys("s", "ctrl+s"),
		key.WithHelp("s", "save & exit"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c", "esc"),
		key.WithHelp("q/esc", "quit without saving"),
	),
	Switch: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch panel"),
	),
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
}

func (k setupKeys) ShortHelp() []key.Binding {
	return []key.Binding{k.Add, k.Remove, k.Switch, k.Save, k.Quit}
}
func (k setupKeys) FullHelp() [][]key.Binding { return nil }

// ── styles ────────────────────────────────────────────────────────────────────

var (
	titleStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).PaddingBottom(1)
	panelStyle     = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
	activePanelSty = panelStyle.BorderForeground(lipgloss.Color("205"))
	dimPanelSty    = panelStyle.BorderForeground(lipgloss.Color("240"))
	selectedSty    = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	dimSty         = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	addedSty       = lipgloss.NewStyle().Foreground(lipgloss.Color("78"))
	hintStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
)

// ── panel enum ────────────────────────────────────────────────────────────────

type panel int

const (
	panelPicker panel = iota
	panelList
)

// ── model ─────────────────────────────────────────────────────────────────────

type setupModel struct {
	cfgPath string
	cfg     config.Config

	fp      filepicker.Model
	help    help.Model
	active  panel
	roots   []string // current scan_roots selection
	listSel int      // selected index in the roots list

	width  int
	height int
	saved  bool
	err    error
}

func newSetupModel(cfgPath string, cfg config.Config) (setupModel, error) {
	fp := filepicker.New()
	fp.DirAllowed = true
	fp.FileAllowed = false
	fp.ShowHidden = false
	fp.CurrentDirectory, _ = os.UserHomeDir()

	roots := make([]string, len(cfg.ScanRoots))
	copy(roots, cfg.ScanRoots)

	return setupModel{
		cfgPath: cfgPath,
		cfg:     cfg,
		fp:      fp,
		help:    help.New(),
		active:  panelPicker,
		roots:   roots,
	}, nil
}

func (m setupModel) Init() tea.Cmd {
	return m.fp.Init()
}

func (m setupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Give picker roughly the left half, minus borders+padding.
		pickerW := msg.Width/2 - 4
		m.fp.Height = msg.Height - 10
		_ = pickerW
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, keys.Save):
			m.cfg.ScanRoots = m.roots
			if err := config.Save(m.cfgPath, m.cfg); err != nil {
				m.err = err
				return m, tea.Quit
			}
			m.saved = true
			return m, tea.Quit

		case key.Matches(msg, keys.Switch):
			if m.active == panelPicker {
				m.active = panelList
			} else {
				m.active = panelPicker
			}
			return m, nil

		// List-panel navigation
		case key.Matches(msg, keys.Up) && m.active == panelList:
			if m.listSel > 0 {
				m.listSel--
			}
			return m, nil

		case key.Matches(msg, keys.Down) && m.active == panelList:
			if m.listSel < len(m.roots)-1 {
				m.listSel++
			}
			return m, nil

		case key.Matches(msg, keys.Remove) && m.active == panelList:
			if len(m.roots) > 0 {
				m.roots = append(m.roots[:m.listSel], m.roots[m.listSel+1:]...)
				if m.listSel >= len(m.roots) && m.listSel > 0 {
					m.listSel--
				}
			}
			return m, nil
		}
	}

	// Delegate to filepicker when picker panel is active.
	if m.active == panelPicker {
		var cmd tea.Cmd
		m.fp, cmd = m.fp.Update(msg)

		if didSelect, path := m.fp.DidSelectFile(msg); didSelect {
			// bubbles filepicker calls directories "files" when DirAllowed=true
			m.roots = appendUnique(m.roots, path)
			m.listSel = len(m.roots) - 1
		}
		return m, cmd
	}

	return m, nil
}

func (m setupModel) View() string {
	if m.width == 0 {
		return "Loading…"
	}

	title := titleStyle.Render("nudged setup — scan_roots")

	// ── left: file picker ──
	pickerBox := m.fp.View()
	pickerLabel := "Browse directories (enter to add)"
	if m.active == panelPicker {
		pickerBox = activePanelSty.Width(m.width/2 - 4).Render(pickerLabel + "\n\n" + pickerBox)
	} else {
		pickerBox = dimPanelSty.Width(m.width/2 - 4).Render(dimSty.Render(pickerLabel) + "\n\n" + pickerBox)
	}

	// ── right: selected list ──
	listLines := []string{}
	if len(m.roots) == 0 {
		listLines = append(listLines, hintStyle.Render("(no directories selected yet)"))
	}
	for i, r := range m.roots {
		line := addedSty.Render("● ") + r
		if m.active == panelList && i == m.listSel {
			line = selectedSty.Render("▶ " + r)
		}
		listLines = append(listLines, line)
	}
	listContent := "Selected scan roots\n\n" + strings.Join(listLines, "\n")
	listW := m.width - m.width/2 - 4
	var listBox string
	if m.active == panelList {
		listBox = activePanelSty.Width(listW).Render(listContent)
	} else {
		listBox = dimPanelSty.Width(listW).Render(dimSty.Render("Selected scan roots") + "\n\n" + strings.Join(listLines, "\n"))
	}

	panels := lipgloss.JoinHorizontal(lipgloss.Top, pickerBox, "  ", listBox)

	helpView := m.help.View(keys)
	var errLine string
	if m.err != nil {
		errLine = "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("Error: "+m.err.Error())
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		panels,
		"",
		helpView,
		errLine,
	)
}

// appendUnique appends s to slice only if it isn't already present.
func appendUnique(slice []string, s string) []string {
	for _, v := range slice {
		if v == s {
			return slice
		}
	}
	return append(slice, s)
}
