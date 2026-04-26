FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -o deadmanswitch ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata su-exec

WORKDIR /app

RUN mkdir -p /app/data

COPY --from=builder /app/deadmanswitch /app/
COPY --from=builder /app/web/templates /app/web/templates
COPY --from=builder /app/web/static /app/web/static
COPY --from=builder /app/internal/email/templates /app/internal/email/templates
COPY entrypoint.sh /entrypoint.sh

RUN chmod +x /entrypoint.sh && \
    addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /app

EXPOSE 8080

ENTRYPOINT ["/entrypoint.sh"]
CMD ["/app/deadmanswitch"]
