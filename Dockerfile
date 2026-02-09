FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download -x

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /app/bin/app ./cmd/app/main.go

FROM alpine:3.20

RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    && addgroup -g 1000 app \
    && adduser -u 1000 -G app -s /bin/sh -D app

WORKDIR /app
COPY --from=builder /app/bin/app /app/bin/app
COPY --from=builder /app/configs /app/configs

RUN chown -R app:app /app

USER app

EXPOSE 8080

CMD ["/app/bin/app"]
