# Build the binary
FROM registry.hub.docker.com/library/golang:1.23 as builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.sum ./

# Copy the go source
COPY . .

RUN go mod tidy
RUN go mod vendor
RUN go build

ENTRYPOINT ["./emulated_workload_operator"]
