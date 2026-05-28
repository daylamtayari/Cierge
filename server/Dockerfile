# Build frontend
FROM --platform=$BUILDPLATFORM node:26-alpine AS frontend
WORKDIR /build
RUN npm install -g pnpm
COPY server/web/package.json server/web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY server/web/ ./
RUN pnpm build

# Build backend
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev

WORKDIR /build/server

RUN apk add --no-cache git

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ ./
COPY --from=frontend /build/dist ./web/dist

RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-X main.Version=${VERSION}" -o /build/bin/server .

# Runtime
FROM alpine:3.23
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /build/bin/server /cierge

EXPOSE 8080
ENTRYPOINT ["/cierge"]
