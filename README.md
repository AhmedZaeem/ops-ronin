# 🥷 Ops-Ronin - Universal TUI Engine

A beautiful and powerful Terminal User Interface (TUI) for managing Docker container operations. Execute commands across multiple containers with ease using an intuitive menu-driven interface.

[![Docker Hub](https://img.shields.io/docker/v/tdkps/ops-ronin?label=Docker%20Hub)](https://hub.docker.com/r/tdkps/ops-ronin)
[![Docker Pulls](https://img.shields.io/docker/pulls/tdkps/ops-ronin)](https://hub.docker.com/r/tdkps/ops-ronin)
[![Build Status](https://github.com/AhmedZaeem/ops-ronin/actions/workflows/deploy.yml/badge.svg)](https://github.com/AhmedZaeem/ops-ronin/actions)
[![Go Version](https://img.shields.io/github/go-mod/go-version/AhmedZaeem/ops-ronin)](https://github.com/AhmedZaeem/ops-ronin)
[![License](https://img.shields.io/github/license/AhmedZaeem/ops-ronin)](LICENSE)

## 🚀 Quick Start with Docker

> **🤖 Automated Updates**: This Docker image is automatically built and deployed from the main branch. Every merge triggers a new build with the latest features and security patches.

### Pull and Run
```bash
docker pull tdkps/ops-ronin:latest
docker run -it --rm -v /var/run/docker.sock:/var/run/docker.sock -v $(pwd)/menu.yaml:/root/menu.yaml tdkps/ops-ronin:latest
```

### Using Docker Compose
```yaml
version: '3.8'
services:
  ops-ronin:
    image: tdkps/ops-ronin:latest
    stdin_open: true
    tty: true
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - ./menu.yaml:/root/menu.yaml
```

## 🚀 Features

- **🖥️ Beautiful TUI**: Clean, intuitive interface built with Charm's Bubble Tea
- **🐳 Docker Integration**: Execute commands in any Docker container
- **📋 Menu-Driven**: Define custom commands in a simple YAML configuration
- **⚡ Real-time Feedback**: See command output and execution status instantly
- **🛡️ Error Handling**: Clear error messages and helpful troubleshooting tips
- **📖 Built-in Help**: Press 'h' for help and keyboard shortcuts

## 📝 Configuration

Create a `menu.yaml` file to define your command menu:

```yaml
project: "my-awesome-project"
theme: "ronin"

menu:
  - title: "Database Operations"
    items:
      - label: "Check DB Status"
        container: "my-database"
        command: "pg_isready -h localhost -p 5432"
      
      - label: "Show Tables"
        container: "my-database"  
        command: "psql -U user -c '\\dt'"

  - title: "App Operations"
    items:
      - label: "Health Check"
        container: "my-app"
        command: "curl -f http://localhost:8080/health"
        
      - label: "View Logs"
        container: "my-app"
        command: "tail -n 20 /var/log/app.log"
```

### Configuration Fields

| Field | Description | Required |
|-------|-------------|----------|
| `project` | Your project name (displayed in the header) | ✅ |
| `theme` | UI theme (currently supports "ronin") | ❌ |
| `menu` | Array of menu categories | ✅ |
| `menu[].title` | Category title | ✅ |
| `menu[].items` | Array of commands in this category | ✅ |
| `menu[].items[].label` | Display name for the command | ✅ |
| `menu[].items[].container` | Target Docker container name | ✅ |
| `menu[].items[].command` | Shell command to execute | ✅ |

## 🎮 Usage

### Keyboard Controls

| Key | Action |
|-----|---------|
| `↑` or `k` | Move up |
| `↓` or `j` | Move down |
| `Enter` | Execute selected command |
| `h` | Toggle help |
| `q` or `Ctrl+C` | Quit |

### Running Commands

1. Use arrow keys or `j`/`k` to navigate the menu
2. Press `Enter` to execute the selected command
3. View the output in the "Last Output" section
4. Green messages indicate success, red messages indicate errors

## 🐳 Docker Setup Examples

### Basic Example
```bash
# Start your application containers
docker run -d --name my-app nginx:alpine
docker run -d --name my-db postgres:15-alpine

# Create menu.yaml
cat > menu.yaml << 'EOF'
project: "my-project"
theme: "ronin"

menu:
  - title: "Web Server"
    items:
      - label: "Check Nginx Status"
        container: "my-app"
        command: "nginx -t"
        
      - label: "Show Access Logs"
        container: "my-app"
        command: "tail -n 10 /var/log/nginx/access.log"

  - title: "Database"
    items:
      - label: "Check DB Connection"
        container: "my-db"
        command: "pg_isready"
EOF

# Run Ops-Ronin
docker run -it --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $(pwd)/menu.yaml:/root/menu.yaml \
  tdkps/ops-ronin:latest
```

### Development Environment
```yaml
version: '3.8'
services:
  database:
    image: postgres:15-alpine
    container_name: my-database
    environment:
      POSTGRES_DB: myapp
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
    volumes:
      - postgres_data:/var/lib/postgresql/data

  backend:
    image: node:18-alpine
    container_name: my-backend
    working_dir: /app
    command: tail -f /dev/null
    volumes:
      - ./backend:/app

  ops-ronin:
    image: tdkps/ops-ronin:latest
    container_name: ops-ronin
    stdin_open: true
    tty: true
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - ./menu.yaml:/root/menu.yaml
    depends_on:
      - database
      - backend

volumes:
  postgres_data:
```

## 🔧 Common Issues & Solutions

### "Container not found" Error
```bash
# Check running containers
docker ps

# Start your container
docker start my-container
```

### "menu.yaml not found" Error
```bash
# Create a basic menu.yaml
cat > menu.yaml << 'EOF'
project: "my-project"
theme: "ronin"
menu:
  - title: "System"
    items:
      - label: "System Info"
        container: "my-container"
        command: "uname -a"
EOF
```

### "Docker not accessible" Error
```bash
# Ensure Docker socket is mounted
docker run -it --rm -v /var/run/docker.sock:/var/run/docker.sock tdkps/ops-ronin:latest

# Check Docker access
docker version
```

## 🛠️ Building from Source

```bash
# Clone the repository
git clone https://github.com/AhmedZaeem/ops-ronin.git
cd ops-ronin

# Build
go build -o ops-ronin main.go

# Run
./ops-ronin
```

### Build with Docker
```bash
# Build image
docker build -t ops-ronin .

# Run
docker run -it --rm -v /var/run/docker.sock:/var/run/docker.sock -v $(pwd)/menu.yaml:/root/menu.yaml ops-ronin
```

## 📚 Examples

### Database Operations
```yaml
- title: "Database Management"
  items:
    - label: "Backup Database"
      container: "my-database"
      command: "pg_dump -U user mydb > /backups/backup.sql"
    
    - label: "Show Database Size"
      container: "my-database"
      command: "psql -U user -c \"SELECT pg_size_pretty(pg_database_size('mydb'));\""
```

### Application Management
```yaml
- title: "Application Operations"
  items:
    - label: "Deploy New Version"
      container: "my-app"
      command: "/scripts/deploy.sh"
    
    - label: "Check Health"
      container: "my-app"
      command: "curl -f http://localhost:8080/health"
```

### System Monitoring
```yaml
- title: "System Monitoring"
  items:
    - label: "Resource Usage"
      container: "my-app"
      command: "top -bn1 | head -20"
    
    - label: "Disk Space"
      container: "my-app"
      command: "df -h"
```

## 🎯 Use Cases

- **DevOps Teams**: Database operations, application deployment, system monitoring
- **Development Teams**: Local environment management, testing, debugging
- **System Administrators**: Infrastructure monitoring, backup operations, maintenance

## 🆘 Need Help?

- **In-app help**: Press `h` while running Ops-Ronin
- **Container issues**: Check `docker ps` and `docker logs <container>`
- **Configuration**: Validate your `menu.yaml` syntax

## 📄 License

This project is licensed under the MIT License. See LICENSE file for details.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

---

**Happy Container Managing! 🥷**
