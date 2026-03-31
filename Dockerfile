FROM golang:1.22-alpine AS builder

WORKDIR /src

# Install certificates for dependency download and HTTPS calls.
RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build a static Linux binary for a small runtime image.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /app/namespace-manager ./cmd/api

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/namespace-manager /app/namespace-manager
COPY --from=builder /src/token_public.pem /app/token_public.pem

EXPOSE 8080

CMD ["/app/namespace-manager"]
