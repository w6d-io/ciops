# Build the dlm binary
FROM golang:1.21.5 as builder
ARG TARGETOS
ARG TARGETARCH
ARG VERSION

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY cmd/ciops.go cmd/ciops.go
COPY api/ api/
COPY internal/ internal/
COPY controllers/ controllers/

# Build
# the GOARCH has not a default value to allow the binary be built according to the host where the command
# was called. For example, if we call make docker-build in a local env which has the Apple Silicon M1 SO
# the docker BUILDPLATFORM arg will be linux/arm64 when for Apple x86 it will be linux/amd64. Therefore,
# by leaving it empty we can ensure that the container and binary shipped on it will have the same platform.
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -ldflags="-X 'github.com/w6d-io/ciops/internal/config.Version=${VERSION}'" -a -o ciops .

# Use distroless as minimal base image to package the ciops binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/ciops .
USER 65532:65532

ENTRYPOINT ["/ciops"]
