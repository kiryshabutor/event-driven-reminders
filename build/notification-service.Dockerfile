FROM golang:alpine AS builder

WORKDIR /src

COPY shared/go.mod shared/go.sum ./shared/
COPY services/notification/go.mod services/notification/go.sum ./services/notification/
RUN cd services/notification && go mod download

COPY shared ./shared
COPY services/notification ./services/notification

RUN cd services/notification && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/notification-service ./cmd

FROM alpine:3.19

WORKDIR /app
COPY --from=builder /out/notification-service .

CMD ["./notification-service"]
