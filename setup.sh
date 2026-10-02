#!/bin/bash
set -e

echo "📦 Inisialisasi backend (Golang)..."
cd backend
go mod init bank-data/backend
go get github.com/gofiber/fiber/v2
go get github.com/go-sql-driver/mysql
go get github.com/golang-migrate/migrate/v4
go get github.com/xuri/excelize/v2
go get github.com/joho/godotenv
cd ..

echo "📦 Inisialisasi frontend (Nuxt 3)..."
cd frontend
npx nuxi@latest init . --force
npm install -D tailwindcss postcss autoprefixer @nuxtjs/tailwindcss
npm install @pinia/nuxt pinia
cd ..

echo "✅ Dependency backend & frontend terpasang."
