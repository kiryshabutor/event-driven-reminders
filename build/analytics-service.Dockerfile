FROM golang:alpine AS builder

WORKDIR /src

COPY shared/go.mod shared/go.sum ./shared/
COPY services/analytics/go.mod services/analytics/go.sum ./services/analytics/
RUN cd services/analytics && go mod download

COPY shared ./shared
COPY services/analytics ./services/analytics

RUN cd services/analytics && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/analytics-service ./cmd

FROM alpine:3.19

RUN apk --no-cache add ca-certificates
RUN adduser -D -g '' appuser

WORKDIR /app
COPY --from=builder /out/analytics-service .
COPY --from=builder /src/services/analytics/app/db/pg/migrations ./migrations

RUN chown -R appuser:appuser /app
USER appuser

EXPOSE 50053
CMD ["./analytics-service"]
