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
ARG SERVICE_NAME=input-service
RUN go build -tags dynamic -o /server cmd/${SERVICE_NAME}/main.go

# Runtime Stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache librdkafka

WORKDIR /root/
COPY --from=builder /server .

# Expose HTTP Port
EXPOSE 8080
EXPOSE 50052

CMD ["./server"]
