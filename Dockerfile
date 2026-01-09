# Go Multistage Build
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies (git for modules, build-base for CGO, librdkafka-dev for dynamic linking)
RUN apk add --no-cache git build-base librdkafka-dev pkgconf

# Copy Mod files
COPY go.mod go.sum ./

# Copy Source
COPY . .

# Install Protoc & Plugins (FIRST)
RUN apk add --no-cache protobuf
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
ENV PATH="$PATH:$(go env GOPATH)/bin"

# Generate Protobufs (Force update)
# Use module=jiaa-server-core to strip the module prefix from the output path, so it matches the expected package text
RUN protoc -I=. --go_out=. --go_opt=module=jiaa-server-core --go-grpc_out=. --go-grpc_opt=module=jiaa-server-core api/proto/core.proto
RUN ls -R pkg/proto/

# Tidy dependencies
RUN go mod tidy -e && go mod vendor

# Build Argument to select service (default: input-service)
# Now building BOTH services for a unified image
RUN go build -tags dynamic -o /server-input cmd/input-service/main.go
RUN go build -tags dynamic -o /server-output cmd/output-service/main.go

# Runtime Stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache librdkafka

WORKDIR /app
# Copy both binaries
COPY --from=builder /server-input ./input-service
COPY --from=builder /server-output ./output-service

# Expose ports for both services
# Input Service: HTTP 8080, gRPC 50052
EXPOSE 8080
EXPOSE 50052
# Output Service: (Assumed port, verify code if needed, but exposing doesn't hurt)

# Default command (can be overridden in k8s)
CMD ["./input-service"]
