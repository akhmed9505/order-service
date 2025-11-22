FROM golang:1.24.10-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o order_service ./cmd/server/main.go

FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/order_service ./order_service
COPY --from=builder /app/.env ./.env
COPY --from=builder /app/frontend ./frontend

EXPOSE 8080

CMD ["./order_service"]
