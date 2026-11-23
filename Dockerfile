# syntax=docker/dockerfile:1

# Stage 1: Build Go backend (dependências vendorizadas, sem rede)
FROM golang:1.27-alpine AS go-builder
WORKDIR /src
COPY backend/ ./
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -mod=vendor -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

# Stage 2: Build frontend (Vue SPA)
FROM node:24-alpine AS node-builder
ENV NODE_OPTIONS="--dns-result-order=ipv4first --no-network-family-autoselection"
RUN corepack enable
WORKDIR /src
COPY frontend/package.json frontend/pnpm-lock.yaml ./frontend/
RUN --mount=type=cache,target=/root/.cache/node/corepack \
    --mount=type=cache,target=/root/.local/share/pnpm/store \
    cd frontend && pnpm install --frozen-lockfile --network-concurrency=4
COPY design-system/tokens.css ./design-system/tokens.css
COPY frontend/ ./frontend/
RUN --mount=type=cache,target=/root/.cache/node/corepack \
    cd frontend && pnpm build

# Stage 3: Runtime - nginx serve SPA e faz proxy /api para Go local
FROM nginx:alpine
RUN apk add --no-cache ca-certificates gettext wget

COPY --from=go-builder /out/api /app/api
COPY --from=node-builder /src/frontend/dist /usr/share/nginx/html
COPY deploy/nginx.conf.template /etc/nginx/templates/default.conf.template
COPY deploy/docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

# Render injeta PORT; API_PORT é interno no container.
ENV PORT=8080
ENV API_PORT=3000

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=10s --retries=5 \
  CMD wget -qO- "http://127.0.0.1:${PORT:-8080}/healthz" >/dev/null || exit 1

CMD ["/docker-entrypoint.sh"]
