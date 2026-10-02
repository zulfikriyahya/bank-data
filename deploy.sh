#!/bin/bash
set -e

echo "🔄 Build backend..."
cd backend
go build -o bin/server ./cmd/server
cd ..

echo "🔄 Build frontend..."
cd frontend
npm run build
cd ..

echo "🔁 Restart PM2..."
pm2 reload ecosystem.config.js --update-env

echo "✅ Deploy selesai."
