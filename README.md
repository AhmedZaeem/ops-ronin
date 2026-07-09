# Ops Ronin V2

<p align="center">
  <img src="assets/ronin.svg" alt="Ops Ronin" width="180" height="180">
</p>

[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.26-00ADD8?logo=go)](https://go.dev)
[![Docker Image](https://img.shields.io/docker/v/tdkps/ops-ronin?sort=semver)](https://hub.docker.com/r/tdkps/ops-ronin)
[![CI](https://github.com/AhmedZaeem/ops-ronin/actions/workflows/docker-publish.yml/badge.svg)](https://github.com/AhmedZaeem/ops-ronin/actions/workflows/docker-publish.yml)

A universal, secure, and fast TUI engine for managing Docker containers through a declarative YAML configuration.

## Why Ops Ronin?

Ops Ronin turns your container operations into a polished keyboard-driven interface. Instead of memorizing long `docker` commands, define them once in `menu.yaml` and navigate with arrow keys, fuzzy search, confirmation dialogs, and live dashboards.

## Features

- **Startup health check**: Verifies configured containers exist and are running before the menu loads.
- **Auto-fix dialog**: Start stopped containers or pick a similar existing container when one is missing.
- **Admin panel**: Browse all containers, images, volumes, and networks; run system prune actions.
- **Declarative menus**: Define containers and commands in YAML.
- **Fuzzy search**: Press `/` to filter commands instantly.
- **Async execution**: Docker tasks run without freezing the UI.
- **Confirmation dialogs**: Destructive actions require explicit approval.
- **Rich output viewer**: Scrollable viewport for logs, status, and command output with JSON syntax highlighting.
- **Live log streaming**: Real-time `docker logs -f` with color-coded log levels.
- **Live resource monitor**: CPU/memory progress bars and ASCII sparklines via the Docker Stats API.
- **Auto-reconnect**: Retry logic for transient Docker daemon failures.
- **Security first**: Input sanitization, strict validation, no raw shell execution.
- **Official Docker SDK**: Uses `github.com/docker/docker/client` instead of spawning CLI processes.
- **Distroless runtime**: Runs as a non-root user on `gcr.io/distroless/static-debian12:nonroot`.
- **Vulnerability scanning**: Every image is scanned with Trivy before it is pushed.

## Installation

### From source

```bash
git clone https://github.com/AhmedZaeem/ops-ronin.git
cd ops-ronin
go build -o ops-ronin ./cmd/ops-ronin
./ops-ronin --config menu.yaml
```

### From Docker

```bash
docker run -it --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $(pwd)/menu.yaml:/app/menu.yaml \
  tdkps/ops-ronin:latest
```

### go install

```bash
go install github.com/AhmedZaeem/ops-ronin/cmd/ops-ronin@latest
```

## Configuration

Create a `menu.yaml` file. The `name` values under `containers` must match actual Docker container names on your host. The startup health check will warn you if any are missing or stopped.

```yaml
title: Production Docker
socket: /var/run/docker.sock
containers:
  - name: nginx
    label: Web Server
    commands:
      - name: logs
        label: View Logs
        action: logs
        args: ["--tail", "100"]

      - name: live-logs
        label: Live Logs
        action: logs-follow
        args: ["--tail", "100"]

      - name: shell
        label: Open Shell
        action: exec
        command: sh

      - name: status
        label: Inspect JSON
        action: status

      - name: monitor
        label: Resource Monitor
        action: monitor

      - name: restart
        label: Restart
        action: restart
        dangerous: true

      - name: stop
        label: Stop
        action: stop
        dangerous: true

  - name: postgres
    label: Database
    commands:
      - name: logs
        label: View Logs
        action: logs
        args: ["--tail", "50"]

      - name: status
        label: Inspect
        action: status

      - name: shell
        label: psql Shell
        action: exec
        command: psql
        args: ["-U", "postgres"]
```

### Supported actions

| Action          | Description                                         |
|-----------------|-----------------------------------------------------|
| `logs`          | Stream recent container logs                        |
| `logs-follow`   | Stream container logs in real time                  |
| `exec`          | Run a command inside the container                  |
| `restart`       | Restart the container                               |
| `stop`          | Stop the container                                  |
| `start`         | Start the container                                 |
| `status`        | Inspect container JSON                              |
| `remove`        | Remove the container                                |
| `prune`         | Remove stopped containers                           |
| `monitor`       | Live CPU/memory/network dashboard with sparklines   |
| `images`        | List Docker images                                  |
| `volumes`       | List Docker volumes                                 |
| `networks`      | List Docker networks                                |
| `network-inspect`| Inspect a specific network                         |

### Advanced examples

Live resource dashboard:

```yaml
- name: monitor
  label: Resource Monitor
  action: monitor
```

Network inspection:

```yaml
- name: inspect-bridge
  label: Inspect Bridge Network
  action: network-inspect
  args: ["bridge"]
```

## Security

- All identifiers and arguments are sanitized against shell metacharacters and control characters.
- Dangerous keywords (`rm`, `remove`, `drop`, `delete`, `prune`, `destroy`, `kill`, `stop`) trigger confirmation dialogs.
- The Docker socket connection uses explicit timeouts and context cancellation.
- The container image is built with a distroless base and runs as a non-root user.
- The CI pipeline scans every image with Trivy and fails on HIGH/CRITICAL CVEs.
- No `os/exec` calls are used; all Docker operations go through the official Go SDK.

> **Note on dependency CVEs:** The `github.com/docker/docker` Go SDK is pinned to the latest available module version (`v28.5.2`). The upstream advisory lists `v29.3.1` as the patched version for several HIGH-severity CVEs, but that module tag has not yet been published to the Go module proxy. A `.trivyignore` file documents these known issues with remediation notes; remove the entries once the patched SDK is available.

## Controls

| Key           | Action                          |
|---------------|---------------------------------|
| `↑` / `↓`     | Navigate commands               |
| `/`           | Fuzzy search                    |
| `a`           | Open admin panel                |
| `Enter`       | Run selected command / fix      |
| `y` / `n`     | Confirm or cancel               |
| `Tab`         | Switch admin tab                |
| `Esc`         | Go back                         |
| `PgUp/PgDn`   | Scroll output                   |
| `p`           | Pause/resume live logs          |
| `q`           | Quit                            |

## Development

```bash
go mod tidy
go test ./...
go vet ./...
go build ./cmd/ops-ronin
```

## CI/CD

The GitHub Actions workflow (`.github/workflows/docker-publish.yml`) runs on every push to `main` and every semantic version tag:

1. Builds the Go binary in a `golang:1.26-alpine` builder stage.
2. Produces a distroless image with `gcr.io/distroless/static-debian12:nonroot`.
3. Scans the image with Trivy; fails on HIGH/CRITICAL vulnerabilities.
4. Uploads the SARIF report to GitHub Security tab.
5. Pushes semantic tags to Docker Hub.

Configure the following repository secrets:

- `DOCKER_USERNAME`
- `DOCKER_PASSWORD`

## License

MIT
