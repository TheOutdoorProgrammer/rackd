# syntax=docker/dockerfile:1

# 1. Build the React SPA. Arch-independent, so it runs on the native build host.
FROM --platform=$BUILDPLATFORM node:22-alpine AS web
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# 2. Build the Go binary, cross-compiled to the target arch (no emulation needed).
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web /web/dist ./web/dist
ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags "-s -w" -o /boating-accident ./cmd/boating-accident

# 3. Minimal static runtime. distroless ships CA certs (for outbound HTTPS to the
#    spec sources) and runs as a non-root user.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /boating-accident /usr/local/bin/boating-accident
ENV BOAT_DATA_DIR=/data \
    BOAT_ADDR=:8080
EXPOSE 8080
VOLUME ["/data"]
ENTRYPOINT ["/usr/local/bin/boating-accident"]
