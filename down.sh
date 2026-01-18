#!/bin/bash

# 1. Load .env (Copy-paste dari logika up.sh)
if [ -f .env ]; then
  set -o allexport
  source .env
  set +o allexport
else
  echo "File .env tidak ditemukan!"
  exit 1
fi

# 2. Jalankan goose down
# PENTING: 'down' di goose hanya membatalkan 1 migrasi terakhir (terbaru).
echo "Rollback 1 langkah migrasi terakhir..."
goose -dir migrations postgres "$DATABASE_URL" down