FROM golang:1.26.1-bookworm AS builder

WORKDIR /build

RUN apt-get update && apt-get install -y \
    build-essential \
    libeb-dev \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*

# Download dependencies before copying source so this layer is only
# invalidated when go.mod or go.sum change, not on every source edit.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o letmesee .

# Download a static ffmpeg build (no shared library dependencies).
FROM debian:bookworm-slim AS ffmpeg-fetcher

RUN apt-get update && apt-get install -y xz-utils curl \
    && rm -rf /var/lib/apt/lists/*

RUN curl -fsSL https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-amd64-static.tar.xz \
    | tar -xJ --strip-components=1 -C /usr/local/bin --wildcards '*/ffmpeg'

# Stage 3: runtime
FROM debian:bookworm-slim

WORKDIR /letmesee

RUN apt-get update && apt-get install -y \
    libeb16 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /build/letmesee /usr/local/bin/letmesee
COPY --from=ffmpeg-fetcher /usr/local/bin/ffmpeg /usr/local/bin/ffmpeg
COPY static/ /letmesee/static/

RUN useradd -m letmesee
USER letmesee

EXPOSE 8080
ENTRYPOINT ["letmesee"]
