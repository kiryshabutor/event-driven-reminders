FROM golang:alpine AS builder

WORKDIR /src

COPY shared/go.mod shared/go.sum ./shared/
COPY services/reminder/go.mod services/reminder/go.sum ./services/reminder/
RUN cd services/reminder && go mod download

COPY shared ./shared
COPY services/reminder ./services/reminder

RUN cd services/reminder && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/reminder-service ./cmd

FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata
RUN adduser -D -g '' appuser

WORKDIR /app
COPY --from=builder /out/reminder-service .
COPY --from=builder /src/services/reminder/app/db/pg/migrations ./migrations

RUN chown -R appuser:appuser /app
USER appuser

ENV TZ=Europe/Moscow
EXPOSE 50052
CMD ["./reminder-service"]
