package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
)

// ContainerSummary is a reduced, UI-friendly representation of a container.
type ContainerSummary struct {
	ID     string
	Name   string
	State  string
	Status string
	Image  string
}

// LogChunk represents one line/chunk from a streaming log source.
type LogChunk struct {
	Stream string // "stdout" or "stderr"
	Data   string
}

// StatsSample is a normalized data point from the Docker Stats API.
type StatsSample struct {
	Timestamp   time.Time
	CPUUsage    float64 // 0-100+ depending on core count
	MemoryUsage uint64
	MemoryLimit uint64
	MemoryPct   float64 // 0-100
	NetworkRx   uint64
	NetworkTx   uint64
}

// Executor defines all Docker operations used by the TUI.
// Implementations must be safe for concurrent use by the Bubble Tea runtime.
type Executor interface {
	Ping(ctx context.Context) error

	ContainerLogs(ctx context.Context, name string, tail string) (string, error)
	ContainerLogsFollow(ctx context.Context, name string, tail string) <-chan LogChunk
	ContainerExec(ctx context.Context, name string, cmd []string) (string, error)
	ContainerRestart(ctx context.Context, name string) error
	ContainerStop(ctx context.Context, name string) error
	ContainerStart(ctx context.Context, name string) error
	ContainerStatus(ctx context.Context, name string) (string, error)
	ContainerRemove(ctx context.Context, name string) error
	ContainerStats(ctx context.Context, name string) (<-chan StatsSample, error)

	PruneContainers(ctx context.Context) (string, error)
	PruneImages(ctx context.Context) (string, error)
	PruneVolumes(ctx context.Context) (string, error)
	PruneNetworks(ctx context.Context) (string, error)

	ListContainers(ctx context.Context) ([]ContainerSummary, error)
	ListImages(ctx context.Context) ([]string, error)
	ListVolumes(ctx context.Context) ([]string, error)
	ListNetworks(ctx context.Context) ([]string, error)
	InspectNetwork(ctx context.Context, id string) (string, error)

	Info(ctx context.Context) (string, error)
}

// SDKExecutor implements Executor using the official Docker Go SDK.
type SDKExecutor struct {
	cli *client.Client
}

// NewSDKExecutor creates an executor backed by the provided Docker client.
func NewSDKExecutor(cli *client.Client) *SDKExecutor {
	return &SDKExecutor{cli: cli}
}

// Ping verifies connectivity to the Docker daemon.
func (e *SDKExecutor) Ping(ctx context.Context) error {
	_, err := e.cli.Ping(ctx)
	return err
}

// ContainerLogs returns the most recent logs for a container.
func (e *SDKExecutor) ContainerLogs(ctx context.Context, name string, tail string) (string, error) {
	if tail == "" {
		tail = "100"
	}

	opts := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       tail,
		Follow:     false,
		Timestamps: false,
	}

	reader, err := e.cli.ContainerLogs(ctx, name, opts)
	if err != nil {
		return "", fmt.Errorf("container logs: %w", err)
	}
	defer reader.Close()

	return demuxDockerLogs(reader), nil
}

// ContainerLogsFollow streams container logs through a channel until the context
// is cancelled or the stream ends.
func (e *SDKExecutor) ContainerLogsFollow(ctx context.Context, name string, tail string) <-chan LogChunk {
	out := make(chan LogChunk)

	go func() {
		defer close(out)

		if tail == "" {
			tail = "100"
		}

		opts := container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Tail:       tail,
			Follow:     true,
			Timestamps: false,
		}

		reader, err := e.cli.ContainerLogs(ctx, name, opts)
		if err != nil {
			select {
			case <-ctx.Done():
			case out <- LogChunk{Stream: "stderr", Data: fmt.Sprintf("stream error: %v", err)}:
			}
			return
		}
		defer reader.Close()

		streamDockerLogs(ctx, reader, out)
	}()

	return out
}

// ContainerExec runs a command inside a container and returns the combined output.
func (e *SDKExecutor) ContainerExec(ctx context.Context, name string, cmd []string) (string, error) {
	if len(cmd) == 0 {
		return "", fmt.Errorf("no command provided")
	}

	execConfig := container.ExecOptions{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          cmd,
	}

	resp, err := e.cli.ContainerExecCreate(ctx, name, execConfig)
	if err != nil {
		return "", fmt.Errorf("exec create: %w", err)
	}

	attach, err := e.cli.ContainerExecAttach(ctx, resp.ID, container.ExecAttachOptions{})
	if err != nil {
		return "", fmt.Errorf("exec attach: %w", err)
	}
	defer attach.Close()

	return demuxDockerLogs(attach.Reader), nil
}

// ContainerRestart restarts a container with a 30-second timeout.
func (e *SDKExecutor) ContainerRestart(ctx context.Context, name string) error {
	timeout := 30
	return e.cli.ContainerRestart(ctx, name, container.StopOptions{Timeout: &timeout})
}

// ContainerStop stops a container with a 30-second timeout.
func (e *SDKExecutor) ContainerStop(ctx context.Context, name string) error {
	timeout := 30
	return e.cli.ContainerStop(ctx, name, container.StopOptions{Timeout: &timeout})
}

// ContainerStart starts a container.
func (e *SDKExecutor) ContainerStart(ctx context.Context, name string) error {
	return e.cli.ContainerStart(ctx, name, container.StartOptions{})
}

// ContainerStatus returns indented JSON inspect output for a container.
func (e *SDKExecutor) ContainerStatus(ctx context.Context, name string) (string, error) {
	info, err := e.cli.ContainerInspect(ctx, name)
	if err != nil {
		return "", fmt.Errorf("container inspect: %w", err)
	}

	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal status: %w", err)
	}

	return string(data), nil
}

// ContainerRemove removes a container.
func (e *SDKExecutor) ContainerRemove(ctx context.Context, name string) error {
	return e.cli.ContainerRemove(ctx, name, container.RemoveOptions{Force: false})
}

// ContainerStats streams resource utilization samples for a container.
func (e *SDKExecutor) ContainerStats(ctx context.Context, name string) (<-chan StatsSample, error) {
	reader, err := e.cli.ContainerStats(ctx, name, true)
	if err != nil {
		return nil, fmt.Errorf("container stats: %w", err)
	}

	out := make(chan StatsSample)

	go func() {
		defer close(out)
		defer reader.Body.Close()

		decoder := json.NewDecoder(reader.Body)
		var previous container.StatsResponse
		first := true

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			var stats container.StatsResponse
			if err := decoder.Decode(&stats); err != nil {
				if err == io.EOF {
					return
				}
				return
			}

			if !first {
				sample := computeStatsSample(&previous, &stats)
				select {
				case <-ctx.Done():
					return
				case out <- sample:
				}
			}

			previous = stats
			first = false
		}
	}()

	return out, nil
}

// PruneContainers removes stopped containers.
func (e *SDKExecutor) PruneContainers(ctx context.Context) (string, error) {
	report, err := e.cli.ContainersPrune(ctx, filters.Args{})
	if err != nil {
		return "", fmt.Errorf("prune containers: %w", err)
	}

	var ids []string
	for _, id := range report.ContainersDeleted {
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		return "No containers pruned", nil
	}

	return fmt.Sprintf("Pruned %d container(s): %s", len(ids), strings.Join(ids, ", ")), nil
}

// ListContainers returns all containers (running and stopped).
func (e *SDKExecutor) ListContainers(ctx context.Context) ([]ContainerSummary, error) {
	containers, err := e.cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}

	result := make([]ContainerSummary, 0, len(containers))
	for _, c := range containers {
		name := ""
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}
		result = append(result, ContainerSummary{
			ID:     c.ID[:12],
			Name:   name,
			State:  c.State,
			Status: c.Status,
			Image:  c.Image,
		})
	}

	return result, nil
}

// ListImages returns human-readable image references.
func (e *SDKExecutor) ListImages(ctx context.Context) ([]string, error) {
	images, err := e.cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list images: %w", err)
	}

	result := make([]string, 0, len(images))
	for _, img := range images {
		tag := "<none>"
		if len(img.RepoTags) > 0 {
			tag = img.RepoTags[0]
		}
		result = append(result, fmt.Sprintf("%s (%s)", tag, img.ID[:12]))
	}

	return result, nil
}

// ListVolumes returns human-readable volume references.
func (e *SDKExecutor) ListVolumes(ctx context.Context) ([]string, error) {
	volumes, err := e.cli.VolumeList(ctx, volume.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list volumes: %w", err)
	}

	result := make([]string, 0, len(volumes.Volumes))
	for _, v := range volumes.Volumes {
		result = append(result, fmt.Sprintf("%s (%s)", v.Name, v.Driver))
	}

	return result, nil
}

// ListNetworks returns human-readable network references.
func (e *SDKExecutor) ListNetworks(ctx context.Context) ([]string, error) {
	networks, err := e.cli.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list networks: %w", err)
	}

	result := make([]string, 0, len(networks))
	for _, n := range networks {
		result = append(result, fmt.Sprintf("%s (%s)", n.Name, n.ID[:12]))
	}

	return result, nil
}

// InspectNetwork returns indented JSON for a network.
func (e *SDKExecutor) InspectNetwork(ctx context.Context, id string) (string, error) {
	info, err := e.cli.NetworkInspect(ctx, id, network.InspectOptions{})
	if err != nil {
		return "", fmt.Errorf("network inspect: %w", err)
	}

	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal network: %w", err)
	}

	return string(data), nil
}

// PruneImages removes dangling images.
func (e *SDKExecutor) PruneImages(ctx context.Context) (string, error) {
	report, err := e.cli.ImagesPrune(ctx, filters.Args{})
	if err != nil {
		return "", fmt.Errorf("prune images: %w", err)
	}

	if len(report.ImagesDeleted) == 0 {
		return "No images pruned", nil
	}

	return fmt.Sprintf("Pruned %d image(s)", len(report.ImagesDeleted)), nil
}

// PruneVolumes removes unused volumes.
func (e *SDKExecutor) PruneVolumes(ctx context.Context) (string, error) {
	report, err := e.cli.VolumesPrune(ctx, filters.Args{})
	if err != nil {
		return "", fmt.Errorf("prune volumes: %w", err)
	}

	if len(report.VolumesDeleted) == 0 {
		return "No volumes pruned", nil
	}

	return fmt.Sprintf("Pruned %d volume(s): %s", len(report.VolumesDeleted), strings.Join(report.VolumesDeleted, ", ")), nil
}

// PruneNetworks removes unused networks.
func (e *SDKExecutor) PruneNetworks(ctx context.Context) (string, error) {
	report, err := e.cli.NetworksPrune(ctx, filters.Args{})
	if err != nil {
		return "", fmt.Errorf("prune networks: %w", err)
	}

	if len(report.NetworksDeleted) == 0 {
		return "No networks pruned", nil
	}

	return fmt.Sprintf("Pruned %d network(s): %s", len(report.NetworksDeleted), strings.Join(report.NetworksDeleted, ", ")), nil
}

// Info returns indented Docker daemon information.
func (e *SDKExecutor) Info(ctx context.Context) (string, error) {
	info, err := e.cli.Info(ctx)
	if err != nil {
		return "", fmt.Errorf("docker info: %w", err)
	}

	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal info: %w", err)
	}

	return string(data), nil
}

// demuxDockerLogs reads the Docker multiplexed stream and returns the combined text.
func demuxDockerLogs(reader io.Reader) string {
	var out strings.Builder
	buf := bufio.NewReader(reader)

	for {
		header := make([]byte, 8)
		_, err := io.ReadFull(buf, header)
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		size := int64(header[4])<<24 | int64(header[5])<<16 | int64(header[6])<<8 | int64(header[7])
		if size <= 0 {
			continue
		}

		payload := make([]byte, size)
		_, err = io.ReadFull(buf, payload)
		if err != nil {
			break
		}

		out.Write(payload)
	}

	return out.String()
}

// streamDockerLogs demultiplexes a Docker log stream and emits each chunk.
func streamDockerLogs(ctx context.Context, reader io.Reader, out chan<- LogChunk) {
	buf := bufio.NewReader(reader)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		header := make([]byte, 8)
		_, err := io.ReadFull(buf, header)
		if err == io.EOF {
			return
		}
		if err != nil {
			return
		}

		size := int64(header[4])<<24 | int64(header[5])<<16 | int64(header[6])<<8 | int64(header[7])
		if size <= 0 {
			continue
		}

		payload := make([]byte, size)
		_, err = io.ReadFull(buf, payload)
		if err != nil {
			return
		}

		stream := "stdout"
		if header[0] == 2 {
			stream = "stderr"
		}

		select {
		case <-ctx.Done():
			return
		case out <- LogChunk{Stream: stream, Data: string(payload)}:
		}
	}
}

// computeStatsSample calculates CPU and memory percentages between two readings.
func computeStatsSample(previous, current *container.StatsResponse) StatsSample {
	cpuDelta := float64(current.CPUStats.CPUUsage.TotalUsage - previous.CPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(current.CPUStats.SystemUsage - previous.CPUStats.SystemUsage)

	onlineCPUs := uint32(len(current.CPUStats.CPUUsage.PercpuUsage))
	if current.CPUStats.OnlineCPUs > 0 {
		onlineCPUs = current.CPUStats.OnlineCPUs
	}

	cpuPct := 0.0
	if systemDelta > 0 && cpuDelta > 0 && onlineCPUs > 0 {
		cpuPct = (cpuDelta / systemDelta) * float64(onlineCPUs) * 100.0
	}

	memUsage := current.MemoryStats.Usage
	memLimit := current.MemoryStats.Limit
	memPct := 0.0
	if memLimit > 0 {
		memPct = float64(memUsage) / float64(memLimit) * 100.0
	}

	var netRx, netTx uint64
	for _, net := range current.Networks {
		netRx += net.RxBytes
		netTx += net.TxBytes
	}

	return StatsSample{
		Timestamp:   time.Now(),
		CPUUsage:    cpuPct,
		MemoryUsage: memUsage,
		MemoryLimit: memLimit,
		MemoryPct:   memPct,
		NetworkRx:   netRx,
		NetworkTx:   netTx,
	}
}

var _ Executor = (*SDKExecutor)(nil)
