package docker

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/docker/docker/client"
)

func NewClient(socketPath string) (*client.Client, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				dialer := &net.Dialer{
					Timeout:   5 * time.Second,
					KeepAlive: 30 * time.Second,
				}
				return dialer.DialContext(ctx, "unix", socketPath)
			},
		},
		Timeout: 60 * time.Second,
	}

	opts := []client.Opt{
		client.WithAPIVersionNegotiation(),
		client.WithHTTPClient(httpClient),
	}

	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, fmt.Errorf("create docker client: %w", err)
	}

	return cli, nil
}
