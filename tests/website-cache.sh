#!/usr/bin/env bash
# Exercise the deployed nginx configuration, including regex location priority.
set -euo pipefail

repo=$(cd "$(dirname "$0")/.." && pwd)
fixture=$(mktemp -d)
container=""
cleanup() {
  if [ -n "$container" ]; then docker rm -f "$container" >/dev/null; fi
  rm -rf "$fixture"
}
trap cleanup EXIT

bundle=app.0123456789abcdef0123456789abcdef.js
mkdir -p "$fixture/static/leafpress/mermaid"
printf 'home' > "$fixture/index.html"
printf 'missing' > "$fixture/404.html"
printf 'body {}' > "$fixture/style.css"
printf 'bundle' > "$fixture/static/leafpress/$bundle"
printf 'mermaid' > "$fixture/static/leafpress/mermaid/mermaid.min.js"
printf 'image' > "$fixture/photo.png"
chmod -R a+rX "$fixture"

container=$(docker run -d --rm -p 127.0.0.1::8080 \
  -v "$fixture:/usr/share/nginx/html:ro" \
  -v "$repo/website/nginx.conf:/etc/nginx/conf.d/default.conf:ro" \
  nginx:stable-alpine)
port=$(docker port "$container" 8080/tcp | cut -d: -f2)
url="http://127.0.0.1:$port"
ready=false
for attempt in $(seq 1 30); do
  if curl -fsS "$url/" >/dev/null 2>&1; then ready=true; break; fi
  sleep 1
done
if [ "$ready" != true ]; then docker logs "$container"; exit 1; fi

headers=$(curl -fsSI "$url/static/leafpress/$bundle")
printf '%s' "$headers" | grep -qi 'Cache-Control:.*immutable'
printf '%s' "$headers" | grep -qi 'Cache-Control:.*max-age=31536000'

for asset in style.css static/leafpress/mermaid/mermaid.min.js photo.png; do
  headers=$(curl -fsSI "$url/$asset")
  printf '%s' "$headers" | grep -qi 'Cache-Control: no-cache'
  if printf '%s' "$headers" | grep -qiE 'immutable|max-age=31536000'; then
    printf 'Mutable asset has long-lived caching: %s\n' "$asset" >&2
    exit 1
  fi
done

test "$(curl -s -o /dev/null -w '%{http_code}' "$url/missing.css")" = 404
printf 'Website cache header checks passed\n'
