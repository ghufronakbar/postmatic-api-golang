#!/bin/bash

# 1. Cek apakah file .env ada
if [ -f .env ]; then
  # Aktifkan auto-export
  set -o allexport
  # Load isi file .env
  source .env
  # Matikan auto-export
  set +o allexport
else
  echo "File .env tidak ditemukan!"
  exit 1
fi

# 2. Jalankan goose
# (Variable $DATABASE_URL sekarang sudah terisi dari .env)
echo "Menjalankan migrasi ke: $DATABASE_URL"
goose -dir migrations postgres "$DATABASE_URL" up