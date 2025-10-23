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
    -o /bin/bid-projector ./services/bid-projector/cmd/main.go

# ------------------------------------------------------------
# Runtime stage
# ------------------------------------------------------------
FROM alpine:3.20

# Install  ca-certificates, tzdata, curl, and gettext (for envsubst)
# ca cert for outbound HTTPS
# tzdata for time zone math
# curl for container health checks
# gettext for environment variable substitution
# non-root user for later to run the app
RUN apk add --no-cache ca-certificates tzdata curl gettext \
    && adduser -D -u 10001 appuser

# set working dir
WORKDIR /app

# copy compiled binary from builder stage
COPY --from=builder /bin/bid-projector /app/bid-projector

# copy config template and entrypoint script
COPY --from=builder /src/services/bid-projector/config.template.json /app/config.template.json
COPY --from=builder /src/services/bid-projector/docker-entrypoint.sh /app/docker-entrypoint.sh

# Make entrypoint executable and change ownership to appuser (as root before switching user)
RUN chmod +x /app/docker-entrypoint.sh && \
    chown -R appuser:appuser /app

# use non root user to run
USER appuser

# run on port 8080
EXPOSE 8080

# run the binary with entrypoint script
ENTRYPOINT ["/app/docker-entrypoint.sh"]
CMD ["/app/bid-projector"]