#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FRONTEND_DIR="${ROOT_DIR}/frontend"

if ! command -v npm >/dev/null 2>&1; then
  echo "Error: npm command not found. Please install Node.js first."
  exit 1
fi

if [[ ! -f "${FRONTEND_DIR}/.env" && -f "${FRONTEND_DIR}/.env.example" ]]; then
  cp "${FRONTEND_DIR}/.env.example" "${FRONTEND_DIR}/.env"
  echo "[frontend] Created frontend/.env from .env.example"
fi

if [[ ! -d "${FRONTEND_DIR}/node_modules" ]]; then
  echo "[frontend] Installing dependencies..."
  cd "${FRONTEND_DIR}"
  npm install
fi

echo "[frontend] Starting Vite dev server on http://localhost:5173 ..."
cd "${FRONTEND_DIR}"
exec npm run dev
