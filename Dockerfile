# Build the manager binary
FROM golang:1.20 as builder

WORKDIR /github.com/w6d-io/ciops
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY internal/ internal/

# Build
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-X 'github.com/w6d-io/ciops/internal/config.Version=${VERSION}' -X 'github.com/w6d-io/ciops/internal/config.Revision=${VCS_REF}' -X 'github.com/w6d-io/ciops/internal/config.Built=${BUILD_DATE}'" \
    -a -o ciops main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/base:nonroot
WORKDIR /
COPY --from=builder /github.com/w6d-io/ciops/ciops .
USER 1001:1001

ENTRYPOINT ["/ciops"]
