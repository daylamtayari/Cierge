# Build
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

ARG TARGETOS TARGETARCH

WORKDIR /build/server

RUN apk add --no-cache git

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ ./
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /build/bin/server .

# Runtime
FROM alpine:3.23
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /build/bin/server /cierge

EXPOSE 8080
ENTRYPOINT ["/cierge"]
