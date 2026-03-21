#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOG_DIR="${ROOT_DIR}/.logs"
mkdir -p "${LOG_DIR}"

echo "[dev] Starting backend and frontend..."
"${ROOT_DIR}/scripts/start-backend.sh" > "${LOG_DIR}/backend.log" 2>&1 &
BACKEND_PID=$!

"${ROOT_DIR}/scripts/start-frontend.sh" > "${LOG_DIR}/frontend.log" 2>&1 &
FRONTEND_PID=$!

echo "[dev] Backend PID: ${BACKEND_PID}"
echo "[dev] Frontend PID: ${FRONTEND_PID}"
echo "[dev] Logs:"
echo "  - ${LOG_DIR}/backend.log"
echo "  - ${LOG_DIR}/frontend.log"
echo "[dev] Press Ctrl+C to stop both services."

cleanup() {
  echo "[dev] Stopping services..."
  kill "${BACKEND_PID}" "${FRONTEND_PID}" >/dev/null 2>&1 || true
}

trap cleanup INT TERM EXIT

wait -n "${BACKEND_PID}" "${FRONTEND_PID}" || true
