# ── stage 1: build ────────────────────────────────────────────────────────────
FROM golang:1.25.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/webanalyzer ./cmd/server

# ── stage 2: run ──────────────────────────────────────────────────────────────
FROM alpine:3.19

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/bin/webanalyzer .

EXPOSE 8080

ENTRYPOINT ["./webanalyzer"]