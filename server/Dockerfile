# Build
FROM golang:1.25-alpine AS builder

WORKDIR /build/server

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ ./
RUN go build -o /build/bin/server .

# Runtime
FROM alpine:3.23
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /build/bin/server /cierge

EXPOSE 8080
ENTRYPOINT ["cierge"]
