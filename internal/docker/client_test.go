package docker

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	cli, err := NewClient("/var/run/docker.sock")
	if err != nil {
		t.Fatalf("expected client creation, got error: %v", err)
	}
	defer cli.Close()
}
