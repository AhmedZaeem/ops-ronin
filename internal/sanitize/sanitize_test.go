package sanitize

import (
	"testing"
)

func TestIdentifier(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"simple", "web", "web", false},
		{"with dash", "web-server", "web-server", false},
		{"with dot", "web.server", "web.server", false},
		{"with colon", "web:8080", "web:8080", false},
		{"empty", "", "", true},
		{"semicolon", "web;rm", "", true},
		{"pipe", "web|cat", "", true},
		{"backtick", "web`rm`", "", true},
		{"dollar", "web$HOME", "", true},
		{"newline", "web\nrm", "", true},
		{"path traversal", "../etc", "", true},
		{"control char", "web\x00", "", true},
		{"unicode", "web_服务器", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Identifier(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Identifier(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Identifier(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestArgument(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"flag", "--tail", false},
		{"number", "100", false},
		{"path", "/var/log/app.log", false},
		{"empty", "", true},
		{"semicolon", "100;rm", true},
		{"ampersand", "100 && rm", true},
		{"pipe", "100 | cat", true},
		{"substitution", "$(id)", true},
		{"backticks", "`id`", true},
		{"redirect", ">/etc/passwd", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Argument(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Argument(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestContainerName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "web", false},
		{"valid with underscore", "web_server", false},
		{"starts with number", "1web", false},
		{"empty", "", true},
		{"starts with dot", ".web", true},
		{"with slash", "web/server", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ContainerName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ContainerName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestSocketPath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"default", "/var/run/docker.sock", false},
		{"relative", "var/run/docker.sock", true},
		{"traversal", "/../var/run/docker.sock", true},
		{"empty", "", true},
		{"null byte", "/var/run/docker.sock\x00", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SocketPath(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("SocketPath(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
