# ------------------------------------------------------------
# Builder stage
# ------------------------------------------------------------
FROM golang:1.24-alpine AS builder
LABEL maintainer="kei"

# use separate working dir
WORKDIR /src

# install git for fetching private deps
RUN apk add --no-cache git

# copy mod files (docker cache)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest
COPY . .

# build for linux/amd64
# - CGO_ENABLED=0 for static linking
# - -ldflags="-s -w" strips symbol/debug info to reduce size
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" \
    -o /bin/bid-command ./services/bid-command/cmd/main.go

# ------------------------------------------------------------
# Runtime stage
# ------------------------------------------------------------
FROM alpine:3.20

# Install  ca-certificates, tzdata and curl
# ca cert for outbound HTTPS
# tzdata for time zone math
# curl for container health checks
# non-root user for later to run the app
RUN apk add --no-cache ca-certificates tzdata curl \
    && adduser -D -u 10001 appuser

# set working dir
WORKDIR /app

# copy compiled binary from builder stage
COPY --from=builder /bin/bid-command /app/bid-command

# copy config file from builder stage
COPY --from=builder /src/services/bid-command/cmd/config.json /app/config.json

# use non root user to run
USER appuser

# run on port 8080
EXPOSE 8080

# hit /healthz endpoint
HEALTHCHECK --interval=10s --timeout=3s --retries=3 \
  CMD curl -fsS http://127.0.0.1:8080/healthz || exit 1

# run the binary
ENTRYPOINT ["/app/bid-command"]