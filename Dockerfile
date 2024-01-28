# Build the dlm binary
FROM golang:1.21.5 as builder
ARG VERSION

WORKDIR /go/src/github.com/w6d-io/ciops

COPY go.mod go.mod
COPY go.sum go.sum
COPY api/ api/
COPY cmd/ cmd/
COPY internal/ internal/
COPY main.go main.go
RUN go mod download && go mod tidy

ENV GO111MODULE="on" \
    GOOS=linux       \
    GOARCH=amd64     \
    CGO_ENABLED=1

# Build
# the GOARCH has not a default value to allow the binary be built according to the host where the command
# was called. For example, if we call make docker-build in a local env which has the Apple Silicon M1 SO
# the docker BUILDPLATFORM arg will be linux/arm64 when for Apple x86 it will be linux/amd64. Therefore,
# by leaving it empty we can ensure that the container and binary shipped on it will have the same platform.
RUN go build -ldflags="-X 'github.com/w6d-io/ciops/internal/config.Version=${VERSION}'" -a -o /ciops .

FROM cgr.dev/chainguard/glibc-dynamic:latest
WORKDIR /
COPY --from=builder /ciops /
USER 65532:65532

ENTRYPOINT ["/ciops"]
