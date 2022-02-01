FROM --platform=$BUILDPLATFORM harbor-repo.vmware.com/dockerhub-proxy-cache/bitnami/golang:1.17 as builder
ARG TARGETPLATFORM
ARG BUILDPLATFORM
# Base Setup
ARG BINARY_NAME
ARG LDFLAGS
WORKDIR /workspace
# Copy the Go Modules manifests

# Copy `go.mod` for definitions and `go.sum` to invalidate the next layer
# in case of a change in the dependencies
COPY go.mod go.sum ./

# Copy src
COPY . .

# Then build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$(echo $TARGETPLATFORM | cut -d '/' -f 2) GO111MODULE=on go build -ldflags "${LDFLAGS}" -a -o ${BINARY_NAME} cmd/wavefront-collector/main.go

# Copy main binary into a thin image
FROM scratch
ARG BINARY_NAME
WORKDIR /
COPY --from=builder /workspace/${BINARY_NAME} .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /workspace/open_source_licenses.txt .

#   nobody:nobody
USER 65534:65534
ENTRYPOINT ["/wavefront-collector"]
