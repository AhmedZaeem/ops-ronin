# syntax=docker/dockerfile:1

# ------------------------------------------------------------------------------
# Stage 1: Builder
# ------------------------------------------------------------------------------
FROM golang:1.26-alpine AS builder

# Install ca-certificates for HTTPS fetches and git for module resolution.
RUN apk add --no-cache ca-certificates git

WORKDIR /build

# Download modules first to leverage Docker layer caching.
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy the rest of the source code.
COPY . .

# Build a static, stripped binary.
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

RUN go build \
    -a \
    -installsuffix cgo \
    -ldflags="-w -s" \
    -trimpath \
    -o ops-ronin \
    ./cmd/ops-ronin

# ------------------------------------------------------------------------------
# Stage 2: Runtime
# ------------------------------------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot

# Copy the binary and a default menu configuration.
COPY --from=builder --chown=nonroot:nonroot /build/ops-ronin /usr/local/bin/ops-ronin
COPY --from=builder --chown=nonroot:nonroot /build/menu.yaml /app/menu.yaml

# Explicitly run as the non-root distroless user.
USER nonroot:nonroot

WORKDIR /app

ENTRYPOINT ["/usr/local/bin/ops-ronin"]
CMD ["--config", "/app/menu.yaml"]
