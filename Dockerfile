FROM golang:1.23-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Собираем бинарник мигратора из каталога cmd/migrator
RUN CGO_ENABLED=0 go build -o migrator ./cmd/migrator
RUN CGO_ENABLED=0 go build -o server ./cmd/server


FROM ubuntu:22.04

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app

COPY --from=builder /app/migrator .
COPY --from=builder /app/server .

COPY config config
COPY migrations migrations