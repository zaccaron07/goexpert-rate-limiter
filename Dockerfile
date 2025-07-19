FROM golang:alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o rate-limiter ./cmd/server

FROM alpine:3.18 AS final

WORKDIR /app

COPY --from=builder /app/rate-limiter ./rate-limiter

COPY env.example .env

EXPOSE 8080

ENTRYPOINT ["./rate-limiter"] 