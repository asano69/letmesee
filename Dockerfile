FROM golang:1.26.1-bookworm AS builder

WORKDIR /build

RUN apt-get update && apt-get install -y \
    build-essential \
    libeb-dev \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum* ./
RUN go mod download || true

COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o letmesee .

# Stage 2: runtime
FROM debian:bookworm-slim

WORKDIR /letmesee

RUN apt-get update && apt-get install -y \
    libeb16 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /build/letmesee /usr/local/bin/letmesee
COPY static/ /letmesee/static/

RUN useradd -m letmesee
USER letmesee

EXPOSE 3000
ENTRYPOINT ["letmesee"]
