FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ops-ronin main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates docker-cli

WORKDIR /root/

COPY --from=builder /app/ops-ronin .
COPY menu.yaml .

CMD ["./ops-ronin"]
