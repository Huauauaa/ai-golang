#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_DIR="${ROOT_DIR}/backend"

if ! command -v go >/dev/null 2>&1; then
  echo "Error: go command not found. Please install Go first."
  exit 1
fi

if command -v docker >/dev/null 2>&1; then
  echo "[backend] Ensuring MySQL container is running..."
  docker compose -f "${ROOT_DIR}/docker-compose.yml" up -d mysql
else
  echo "[backend] docker not found, skip MySQL auto-start."
fi

if [[ ! -f "${BACKEND_DIR}/.env" && -f "${BACKEND_DIR}/.env.example" ]]; then
  cp "${BACKEND_DIR}/.env.example" "${BACKEND_DIR}/.env"
  echo "[backend] Created backend/.env from .env.example"
fi

echo "[backend] Starting Gin server on http://localhost:8080 ..."
cd "${BACKEND_DIR}"
exec go run .
