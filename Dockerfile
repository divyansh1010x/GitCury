# Multi-stage build for GitCury
FROM golang:1.24.1-alpine AS builder

# Install git and ca-certificates
RUN apk --no-cache add git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o gitcury main.go

# Final stage - minimal image
FROM scratch

# Copy certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary
COPY --from=builder /build/gitcury /usr/local/bin/gitcury

# Copy default config
COPY --from=builder /build/config.json /etc/gitcury/config.json

# Set entrypoint
ENTRYPOINT ["/usr/local/bin/gitcury"]
CMD ["--help"]

# Labels
LABEL maintainer="Lakshya Jain <lakshyajain0291@gmail.com>"
LABEL description="AI-powered Git automation CLI tool"
LABEL version="latest"
LABEL org.opencontainers.image.source="https://github.com/lakshyajain-0291/GitCury"
