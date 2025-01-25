FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o go-w .

FROM alpine:latest

RUN apk add --no-cache libc6-compat

COPY --from=builder /app/go-w /usr/local/bin/go-w

CMD ["go-w"]