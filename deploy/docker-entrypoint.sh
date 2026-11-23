#!/bin/sh
set -eu

PORT=${PORT:-8080}
API_PORT=${API_PORT:-3000}
export PORT API_PORT

envsubst '${PORT} ${API_PORT}' \
  < /etc/nginx/templates/default.conf.template \
  > /etc/nginx/http.d/default.conf

# Go API escuta em localhost; nginx expõe a porta pública do Render.
PORT=$API_PORT /app/api &
api_pid=$!

shutdown() {
  kill "$api_pid" 2>/dev/null || true
  nginx -s quit 2>/dev/null || true
  wait "$api_pid" 2>/dev/null || true
}

trap shutdown TERM INT

nginx -g 'daemon off;' &
nginx_pid=$!

# Se qualquer processo sair, encerra o outro e propaga falha.
while :; do
  if ! kill -0 "$api_pid" 2>/dev/null; then
    wait "$api_pid" || api_status=$?
    kill "$nginx_pid" 2>/dev/null || true
    wait "$nginx_pid" 2>/dev/null || true
    exit "${api_status:-1}"
  fi
  if ! kill -0 "$nginx_pid" 2>/dev/null; then
    wait "$nginx_pid" || nginx_status=$?
    kill "$api_pid" 2>/dev/null || true
    wait "$api_pid" 2>/dev/null || true
    exit "${nginx_status:-1}"
  fi
  sleep 1
done
