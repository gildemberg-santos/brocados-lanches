FROM golang:1.27-alpine AS build

WORKDIR /src

# Go dependencies are vendored; the build does not access the network.
COPY backend/ ./

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -mod=vendor -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM alpine:3.22

RUN addgroup -S app \
    && adduser -S -G app app

WORKDIR /app

COPY --from=build /out/api /app/api

USER app

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=5 \
  CMD wget -qO- http://127.0.0.1:8080/healthz >/dev/null || exit 1

ENTRYPOINT ["/app/api"]
