FROM golang:alpine AS builder

WORKDIR /src

COPY shared/go.mod shared/go.sum ./shared/
COPY services/gateway/go.mod services/gateway/go.sum ./services/gateway/
RUN cd services/gateway && go mod download

COPY shared ./shared
COPY services/gateway ./services/gateway

RUN cd services/gateway && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/api-gateway ./cmd

FROM alpine:3.19

RUN apk --no-cache add ca-certificates
RUN adduser -D -g '' appuser

WORKDIR /app
COPY --from=builder /out/api-gateway .

RUN chown -R appuser:appuser /app
USER appuser

EXPOSE 8080
CMD ["./api-gateway"]
