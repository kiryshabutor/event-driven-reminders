FROM golang:alpine AS builder

WORKDIR /src

COPY shared/go.mod shared/go.sum ./shared/
COPY services/auth/go.mod services/auth/go.sum ./services/auth/
RUN cd services/auth && go mod download

COPY shared ./shared
COPY services/auth ./services/auth

RUN cd services/auth && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/auth-service ./cmd

FROM alpine:3.19

RUN apk --no-cache add ca-certificates
RUN adduser -D -g '' appuser

WORKDIR /app
COPY --from=builder /out/auth-service .
COPY --from=builder /src/services/auth/app/db/pg/migrations ./migrations

RUN chown -R appuser:appuser /app
USER appuser

EXPOSE 50051
CMD ["./auth-service"]
