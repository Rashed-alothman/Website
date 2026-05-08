# ── Build stage ───────────────────────────────────────────────────────────────
# golang:1.26-alpine matches the Go version in go.mod.
# CGO_ENABLED=0 produces a fully static binary — no glibc needed in the runner.
FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

# ── Run stage ─────────────────────────────────────────────────────────────────
# Alpine is ~7 MB. ca-certificates is needed for any future HTTPS outbound calls.
FROM alpine:3.20

RUN apk --no-cache add ca-certificates
WORKDIR /app

COPY --from=builder /app/server     ./server
COPY templates/                     ./templates/
COPY static/                        ./static/

EXPOSE 8080
CMD ["./server"]
