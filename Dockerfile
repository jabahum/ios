# syntax=docker/dockerfile:1

############################
# Builder
############################
FROM golang:1.24 AS builder

WORKDIR /src

# Cache dependencies first
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build static binary
# Change ./cmd/server if your main package is elsewhere
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -ldflags="-s -w" -o /app/integrated-outbreak-system ./cmd/server

############################
# Runtime
############################
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /app/integrated-outbreak-system /app/integrated-outbreak-system

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/app/integrated-outbreak-system"]