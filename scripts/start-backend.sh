#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_DIR="${ROOT_DIR}/backend"
MYSQL_CONTAINER_NAME="${MYSQL_CONTAINER_NAME:-demo-mysql}"

if ! command -v go >/dev/null 2>&1; then
  echo "Error: go command not found. Please install Go first."
  exit 1
fi

if command -v docker >/dev/null 2>&1; then
  if docker container inspect "${MYSQL_CONTAINER_NAME}" >/dev/null 2>&1; then
    if [[ "$(docker inspect -f '{{.State.Running}}' "${MYSQL_CONTAINER_NAME}")" != "true" ]]; then
      echo "[backend] Starting existing MySQL container: ${MYSQL_CONTAINER_NAME}"
      docker start "${MYSQL_CONTAINER_NAME}" >/dev/null
    else
      echo "[backend] Reusing existing MySQL container: ${MYSQL_CONTAINER_NAME}"
    fi
  else
    echo "[backend] Ensuring MySQL container is running..."
    docker compose -f "${ROOT_DIR}/docker-compose.yml" up -d mysql
  fi
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
