# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown

RUN go build \
    -ldflags "-X github.com/nomos/nomos/internal/version.Version=${VERSION} \
              -X github.com/nomos/nomos/internal/version.Commit=${COMMIT} \
              -X github.com/nomos/nomos/internal/version.Date=${DATE} \
              -X github.com/nomos/nomos/internal/version.BuiltBy=docker" \
    -o /nomos ./cmd/nomos

# Runtime stage
FROM alpine:3.20

RUN apk add --no-cache git ca-certificates tzdata

COPY --from=builder /nomos /usr/local/bin/nomos
COPY scripts/create-demo-cosmos.sh /usr/local/bin/create-demo-cosmos.sh

WORKDIR /cosmos

ENV COSMOS_PATH=/cosmos/data
# Base directory for new local repositories. Kept separate from COSMOS_PATH so
# repository creation works even when the cosmos data is a read-only mount.
ENV NOMOS_REPOS_DIR=/cosmos/repos
ENV NOMOS_BIN=/usr/local/bin/nomos
ENV NOMOS_LISTEN=0.0.0.0:7373

EXPOSE 7373

COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh /usr/local/bin/create-demo-cosmos.sh

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
