FROM golang:1.26-alpine AS builder

WORKDIR /src
ENV CGO_ENABLED=0 GOWORK=off

COPY protocol ./protocol
COPY server/go.mod server/go.sum ./server/
WORKDIR /src/server
RUN go mod download

COPY server ./
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown
RUN go build -trimpath \
    -ldflags="-s -w \
      -X github.com/z9wen/VPSentinel/server/internal/version.Version=${VERSION} \
      -X github.com/z9wen/VPSentinel/server/internal/version.Commit=${COMMIT} \
      -X github.com/z9wen/VPSentinel/server/internal/version.Date=${BUILD_DATE}" \
    -o /out/vpsentinel-server \
    ./cmd/vpsentinel-server

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S vpsentinel \
    && adduser -S -G vpsentinel -H vpsentinel

COPY --from=builder /out/vpsentinel-server /usr/local/bin/vpsentinel-server

USER vpsentinel:vpsentinel
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/vpsentinel-server"]
