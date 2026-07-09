package tui

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/AhmedZaeem/ops-ronin/internal/config"
	"github.com/AhmedZaeem/ops-ronin/internal/docker"
	"github.com/AhmedZaeem/ops-ronin/internal/monitoring"
)

// Model is the top-level Bubble Tea model for Ops Ronin.
type Model struct {
	cfg      *config.Config
	executor docker.Executor

	width  int
	height int

	state state

	list        list.Model
	textInput   textinput.Model
	viewport    viewport.Model
	spinner     spinner.Model
	spinnerInit bool

	selectedItem commandItem
	output       string
	err          error
	confirmMsg   string

	healthReport        *healthReport
	healthLoading       bool
	selectedHealthIndex int
	autoFixSuggestions  []docker.ContainerSummary
	autoFixSelected     int

	admin       adminModel
	adminOutput string

	// Live log streaming state.
	logBuffer    string
	logChannel   <-chan docker.LogChunk
	logPaused    bool
	logCancelled context.CancelFunc

	// Live monitoring state.
	statsChannel   <-chan docker.StatsSample
	statsCancelled context.CancelFunc
	cpuHistory     *monitoring.RingBuffer
	memHistory     *monitoring.RingBuffer
	cpuProgress    progress.Model
	memProgress    progress.Model
}

// runFinishedMsg is emitted when a one-shot command completes.
type runFinishedMsg struct {
	output string
}

// runErrorMsg is emitted when a one-shot command fails.
type runErrorMsg struct {
	err error
}

// healthFinishedMsg is emitted after the startup health check.
type healthFinishedMsg struct {
	report *healthReport
	err    error
}

// logChunkMsg delivers a single line/chunk from a streaming log source.
type logChunkMsg struct {
	chunk docker.LogChunk
}

// logDoneMsg is emitted when a log stream closes.
type logDoneMsg struct{}

// statsSampleMsg delivers a single resource utilization sample.
type statsSampleMsg struct {
	sample docker.StatsSample
}

// statsDoneMsg is emitted when a stats stream closes.
type statsDoneMsg struct{}

// NewModel creates the initial TUI model.
func NewModel(cfg *config.Config, executor docker.Executor) *Model {
	items := buildItems(cfg)

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Foreground(lipgloss.Color("#7AA2F7")).Bold(true)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Foreground(lipgloss.Color("#A9B1D6"))

	lst := list.New(items, delegate, 0, 0)
	lst.Title = cfg.Title
	lst.SetShowStatusBar(false)
	lst.SetFilteringEnabled(false)
	lst.Styles.Title = titleStyle
	lst.Styles.HelpStyle = helpStyle

	ti := textinput.New()
	ti.Prompt = "/ "
	ti.Placeholder = "type to filter commands..."
	ti.CharLimit = 64

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = highlightStyle

	vp := viewport.New(0, 0)
	vp.Style = viewportStyle

	cpuBar := progress.New(progress.WithDefaultGradient())
	memBar := progress.New(progress.WithDefaultGradient())

	return &Model{
		cfg:           cfg,
		executor:      executor,
		state:         stateHealth,
		healthLoading: true,
		list:          lst,
		textInput:     ti,
		viewport:      vp,
		spinner:       sp,
		admin:         newAdminModel(0, 0),
		cpuHistory:    monitoring.NewRingBuffer(60),
		memHistory:    monitoring.NewRingBuffer(60),
		cpuProgress:   cpuBar,
		memProgress:   memBar,
	}
}

// Init starts the startup health check and spinner.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			names := make([]string, 0, len(m.cfg.Containers))
			for _, c := range m.cfg.Containers {
				names = append(names, c.Name)
			}
			report, err := checkContainerHealth(ctx, m.executor, names)
			return healthFinishedMsg{report: report, err: err}
		},
	)
}

func (m *Model) setSize(width, height int) {
	m.width = width
	m.height = height

	headerHeight := 4
	footerHeight := 2

	m.list.SetSize(width, height-headerHeight-footerHeight)
	m.viewport.Width = width - 4
	m.viewport.Height = height - headerHeight - footerHeight - 2
	m.textInput.Width = width - 6

	m.admin.SetSize(width, height-headerHeight-footerHeight)

	barWidth := width - 12
	if barWidth < 20 {
		barWidth = 20
	}
	m.cpuProgress.Width = barWidth
	m.memProgress.Width = barWidth
}
