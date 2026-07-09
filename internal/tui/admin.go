package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/AhmedZaeem/ops-ronin/internal/docker"
)

type adminTab int

const (
	adminTabContainers adminTab = iota
	adminTabImages
	adminTabVolumes
	adminTabPrune
)

type adminContainerItem struct {
	summary docker.ContainerSummary
}

func (i adminContainerItem) FilterValue() string {
	return i.summary.Name + " " + i.summary.Image + " " + i.summary.Status
}

func (i adminContainerItem) Title() string {
	return i.summary.Name
}

func (i adminContainerItem) Description() string {
	return fmt.Sprintf("%s  %s", i.summary.Image, i.summary.Status)
}

type adminStringItem struct {
	value string
}

func (i adminStringItem) FilterValue() string { return i.value }
func (i adminStringItem) Title() string       { return i.value }
func (i adminStringItem) Description() string { return "" }

type adminModel struct {
	tab        adminTab
	containers list.Model
	images     list.Model
	volumes    list.Model
	prune      list.Model
	loading    bool
	err        error
}

type adminLoadedMsg struct {
	containers []docker.ContainerSummary
	images     []string
	volumes    []string
	err        error
}

type adminActionMsg struct {
	output string
	err    error
}

func newAdminModel(width, height int) adminModel {
	makeList := func(title string) list.Model {
		l := list.New([]list.Item{}, list.NewDefaultDelegate(), width, height)
		l.Title = title
		l.SetShowStatusBar(false)
		l.SetFilteringEnabled(false)
		l.Styles.Title = titleStyle
		return l
	}

	pruneItems := []list.Item{
		adminStringItem{value: "Prune stopped containers"},
		adminStringItem{value: "Prune dangling images"},
		adminStringItem{value: "Prune unused volumes"},
		adminStringItem{value: "Prune unused networks"},
	}

	pruneList := makeList("System Cleanup")
	pruneList.SetItems(pruneItems)

	return adminModel{
		containers: makeList("Containers"),
		images:     makeList("Images"),
		volumes:    makeList("Volumes"),
		prune:      pruneList,
		loading:    true,
	}
}

func (a *adminModel) SetSize(width, height int) {
	headerFooter := 6
	a.containers.SetSize(width, height-headerFooter)
	a.images.SetSize(width, height-headerFooter)
	a.volumes.SetSize(width, height-headerFooter)
	a.prune.SetSize(width, height-headerFooter)
}

func (a *adminModel) CurrentList() *list.Model {
	switch a.tab {
	case adminTabContainers:
		return &a.containers
	case adminTabImages:
		return &a.images
	case adminTabVolumes:
		return &a.volumes
	default:
		return &a.prune
	}
}

func loadAdminData(ctx context.Context, executor docker.Executor) tea.Msg {
	containers, err := executor.ListContainers(ctx)
	if err != nil {
		return adminLoadedMsg{err: err}
	}

	images, err := executor.ListImages(ctx)
	if err != nil {
		return adminLoadedMsg{err: err}
	}

	volumes, err := executor.ListVolumes(ctx)
	if err != nil {
		return adminLoadedMsg{err: err}
	}

	return adminLoadedMsg{
		containers: containers,
		images:     images,
		volumes:    volumes,
	}
}

func (a *adminModel) UpdateData(msg adminLoadedMsg) {
	a.loading = false
	if msg.err != nil {
		a.err = msg.err
		return
	}

	containerItems := make([]list.Item, 0, len(msg.containers))
	for _, c := range msg.containers {
		containerItems = append(containerItems, adminContainerItem{summary: c})
	}
	a.containers.SetItems(containerItems)

	imageItems := make([]list.Item, 0, len(msg.images))
	for _, img := range msg.images {
		imageItems = append(imageItems, adminStringItem{value: img})
	}
	a.images.SetItems(imageItems)

	volumeItems := make([]list.Item, 0, len(msg.volumes))
	for _, vol := range msg.volumes {
		volumeItems = append(volumeItems, adminStringItem{value: vol})
	}
	a.volumes.SetItems(volumeItems)
}

func (a *adminModel) View(width int) string {
	if a.loading {
		return lipgloss.Place(width, 10, lipgloss.Center, lipgloss.Center, "Loading admin data...")
	}

	if a.err != nil {
		return errorStyle.Render(fmt.Sprintf("Admin error: %v", a.err))
	}

	tabs := []string{"Containers", "Images", "Volumes", "Prune"}
	var renderedTabs []string
	for i, t := range tabs {
		if adminTab(i) == a.tab {
			renderedTabs = append(renderedTabs, highlightStyle.Render("["+t+"]"))
		} else {
			renderedTabs = append(renderedTabs, helpStyle.Render(t))
		}
	}

	tabBar := strings.Join(renderedTabs, "  ")
	content := a.CurrentList().View()

	return lipgloss.JoinVertical(lipgloss.Left, tabBar, content)
}
