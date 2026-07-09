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

# Copy only the source tree required to build the application.
COPY cmd ./cmd
COPY internal ./internal
COPY menu.yaml ./menu.yaml

# Build a static, stripped binary for the target platform.
ARG TARGETOS
ARG TARGETARCH
ENV CGO_ENABLED=0
ENV GOOS=${TARGETOS}
ENV GOARCH=${TARGETARCH}

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

# Copy the binary with explicit executable permissions and immutable ownership.
COPY --from=builder --chown=nonroot:nonroot --chmod=755 /build/ops-ronin /usr/local/bin/ops-ronin

# Copy the default menu configuration as read-only.
COPY --from=builder --chown=nonroot:nonroot --chmod=444 /build/menu.yaml /app/menu.yaml

# Explicitly run as the non-root distroless user (UID 65532).
USER nonroot:nonroot

# Use a dedicated, non-writable application directory.
WORKDIR /app

ENTRYPOINT ["/usr/local/bin/ops-ronin"]
CMD ["--config", "/app/menu.yaml"]
