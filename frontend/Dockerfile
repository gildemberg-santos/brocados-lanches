FROM node:24-alpine AS build

# Keep builds predictable on IPv4-only or dual-stack Docker networks.
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

FROM nginx:alpine

# Defaults para Compose; o Render define PORT e API_HOSTPORT em runtime.
ENV PORT=80
ENV API_HOSTPORT=api:8080

COPY frontend/nginx.conf /etc/nginx/templates/default.conf.template
COPY --from=build /src/frontend/dist /usr/share/nginx/html

EXPOSE 80

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=5 \
  CMD ["sh", "-c", "wget -qO- \"http://127.0.0.1:${PORT:-80}/\" >/dev/null || exit 1"]
