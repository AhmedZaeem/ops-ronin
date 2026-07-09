# Ops Ronin — Universal Secure Docker TUI Engine

> A keyboard-driven, distroless TUI for managing Docker containers, volumes, networks, images, and live resource monitoring — with zero shell execution.

[![GitHub](https://img.shields.io/badge/source-AhmedZaeem%2Fops--ronin-blue?logo=github)](https://github.com/AhmedZaeem/ops-ronin)
[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev)
[![Distroless](https://img.shields.io/badge/runtime-distroless%20nonroot-2ea44f)](https://github.com/GoogleContainerTools/distroless)

---

## What is Ops Ronin?

**Ops Ronin** is a secure, fast, and beautiful Terminal User Interface (TUI) for Docker operations. Instead of memorizing long `docker` commands, define your containers and actions once in a declarative `menu.yaml` file, then navigate with arrow keys, fuzzy search, and live dashboards.

Built for **DevOps engineers, SREs, platform teams, and developers** who want a safer, faster way to interact with Docker daemons.

---

## Why use this image?

- 🎛️ **Keyboard-driven TUI** — arrow keys, `/` search, confirmation dialogs
- 🔒 **100% shell-injection free** — uses the official Docker Go SDK, never `os/exec`
- 📊 **Live monitoring** — CPU, memory, and network sparklines via Docker Stats API
- 📜 **Live log streaming** — real-time `docker logs -f` with color-coded log levels
- 🛡️ **Distroless & non-root** — minimal attack surface, runs as unprivileged user
- 🔍 **Trivy-scanned** — CI fails on HIGH/CRITICAL CVEs before publishing
- ⚡ **Multi-arch** — `linux/amd64` and `linux/arm64` (Apple Silicon ready)

---

## Quick Start

```bash
docker run -it --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $(pwd)/menu.yaml:/app/menu.yaml \
  tdkps/ops-ronin:latest
```

### With a custom config path

```bash
docker run -it --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /path/to/menu.yaml:/app/menu.yaml \
  tdkps/ops-ronin:latest --config /app/menu.yaml
```

---

## Example menu.yaml

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

      - name: monitor
        label: Resource Monitor
        action: monitor

      - name: shell
        label: Open Shell
        action: exec
        command: sh

      - name: restart
        label: Restart
        action: restart
        dangerous: true

      - name: stop
        label: Stop
        action: stop
        dangerous: true
```

### Supported actions

| Action | Description |
|--------|-------------|
| `logs` | View recent container logs |
| `logs-follow` | Stream logs in real time |
| `exec` | Run commands inside containers |
| `status` | Inspect container JSON |
| `monitor` | Live CPU / memory / network dashboard |
| `restart` | Restart container |
| `start` | Start container |
| `stop` | Stop container |
| `remove` | Remove container |
| `prune` | Remove stopped containers |
| `images` | List Docker images |
| `volumes` | List Docker volumes |
| `networks` | List Docker networks |
| `network-inspect` | Inspect a specific network |

---

## Controls

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate commands |
| `/` | Fuzzy search |
| `Enter` | Run selected command |
| `y` / `n` | Confirm or cancel |
| `a` | Open admin panel |
| `Tab` | Switch admin tab |
| `p` | Pause/resume live logs |
| `PgUp` / `PgDn` | Scroll output |
| `Esc` | Go back |
| `q` | Quit |

---

## Available Tags

| Tag | Description |
|-----|-------------|
| `latest` | Latest stable release |
| `2.1.6` | Specific release version |
| `2.1` | Latest patch in the 2.1 minor line |

---

## Security

- Distroless `gcr.io/distroless/static-debian12:nonroot` runtime
- Runs as unprivileged `nonroot` user (UID 65532)
- No shell, no package manager, no writable OS directories
- Input sanitization against shell metacharacters and control characters
- Dangerous keywords trigger confirmation dialogs
- Docker socket connection uses strict timeouts
- Trivy vulnerability scanning in CI pipeline
- Multi-arch builds for `linux/amd64` and `linux/arm64`

---

## Source & Issues

- **GitHub:** https://github.com/AhmedZaeem/ops-ronin
- **License:** MIT

---

# Docker Hub Short Description

**Ops Ronin** — A secure, distroless, keyboard-driven TUI engine for managing Docker containers, logs, volumes, networks, images, and live resource monitoring. Multi-arch (`amd64`/`arm64`), Trivy-scanned, zero shell execution.
