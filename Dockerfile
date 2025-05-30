# Stage 1: Build the Go application
FROM golang:1.24-alpine AS builder
WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the application
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${BUILD_DATE}" -a -installsuffix cgo -o gitcury ./main.go

# Stage 2: Create the final lightweight image
FROM alpine:latest

# Add labels
LABEL org.opencontainers.image.title="GitCury"
LABEL org.opencontainers.image.description="AI-Powered Git Automation CLI tool"
LABEL org.opencontainers.image.url="https://github.com/lakshyajain-0291/GitCury"
LABEL org.opencontainers.image.source="https://github.com/lakshyajain-0291/GitCury"
LABEL org.opencontainers.image.licenses="MIT"

# Install necessary dependencies
RUN apk --no-cache add git bash curl jq

WORKDIR /app/

# Copy the binary and example config from the builder stage
COPY --from=builder /app/gitcury .
COPY --from=builder /app/config.json.example /app/config.json.example

# Make the binary executable before switching users
RUN chmod +x ./gitcury

# Create a non-root user for security
RUN addgroup -S gitcurygroup && adduser -S gitcuryuser -G gitcurygroup

# Set home directory for the user (GitCury might store config in $HOME/.gitcury)
ENV HOME=/home/gitcuryuser
RUN mkdir -p $HOME/.gitcury

# Copy default config to the user's config directory and set proper ownership
COPY --from=builder /app/config.json.example $HOME/.gitcury/config.json.example
RUN chown -R gitcuryuser:gitcurygroup $HOME && chown gitcuryuser:gitcurygroup /app/gitcury

# Create a wrapper script to handle API key from environment
RUN printf '#!/bin/bash\n\
# Copy example config if no config exists\n\
if [ ! -f $HOME/.gitcury/config.json ]; then\n\
  cp $HOME/.gitcury/config.json.example $HOME/.gitcury/config.json\n\
  # Replace placeholder with actual API key if provided\n\
  if [ ! -z "$GEMINI_API_KEY" ]; then\n\
    sed -i "s/YOUR_GEMINI_API_KEY/$GEMINI_API_KEY/g" $HOME/.gitcury/config.json\n\
  fi\n\
fi\n\
# Run the application\n\
/app/gitcury "$@"\n' > /app/entrypoint.sh && chmod +x /app/entrypoint.sh

# Switch to non-root user
USER gitcuryuser

ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["--help"]
