#!/bin/bash
set -e

echo "🔨 Building core-app..."
cd backend
go build -o bin/server ./cmd/server
cd ..

echo "🔁 Restarting PM2..."
pm2 restart core-app

echo "✅ Selesai. Cek status:"
pm2 status
