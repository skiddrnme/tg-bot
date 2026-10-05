FROM golang:1.26-alpine AS builder
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o /out/bot ./cmd/bot

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S app && adduser -S app -G app
WORKDIR /app
COPY --from=builder /out/bot /app/bot
RUN mkdir -p /app/certs && chown -R app:app /app
USER app
EXPOSE 8080
ENTRYPOINT ["/app/bot"]