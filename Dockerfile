FROM golang:1.25.6-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o market ./cmd/webserver

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/market .

EXPOSE 8080

CMD ["./market"]