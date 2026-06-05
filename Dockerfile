# Build stage
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder
WORKDIR /build

# Cross-compilation target (provided by buildx) and version stamping args.
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
ARG COMMIT=none
ARG DATE=unknown

COPY go.mod go.sum ./
RUN go mod download
COPY . .
# CGO is disabled, so GOOS/GOARCH cross-compile natively on the build host -- no QEMU.
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags="-w -s -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
    -o karaclean ./cmd/karaclean

# Final stage
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/karaclean /karaclean
ENTRYPOINT ["/karaclean"]
