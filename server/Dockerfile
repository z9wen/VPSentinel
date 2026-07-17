FROM golang:1.26-alpine AS builder

WORKDIR /src
ENV CGO_ENABLED=0 GOWORK=off

COPY protocol ./protocol
COPY server/go.mod server/go.sum ./server/
WORKDIR /src/server
RUN go mod download

COPY server ./
RUN go build -trimpath -ldflags="-s -w" -o /out/vpsentinel-server ./cmd/vpsentinel-server

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S vpsentinel \
    && adduser -S -G vpsentinel -H vpsentinel

COPY --from=builder /out/vpsentinel-server /usr/local/bin/vpsentinel-server

USER vpsentinel:vpsentinel
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/vpsentinel-server"]
