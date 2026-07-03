#!/bin/bash
set -e

echo "==========================================="
echo "   SIAKAD NAWAWI - AUTOMATED BUILD SCRIPT  "
echo "==========================================="

# 1. Build Backend (Golang)
echo "[1/3] Building Backend (Golang)..."
cd backend
go build -o server main.go
echo "✅ Backend build success (backend/server)"
cd ..

# 2. Build Frontend (SvelteKit)
echo "[2/3] Building Frontend (SvelteKit)..."
cd frontend
npm install
npm run build
echo "✅ Frontend build success (frontend/build or frontend/dist)"
cd ..

# 3. Build Mobile (Svelte SPA)
echo "[3/3] Building Mobile (Svelte SPA)..."
cd mobile
npm install
npm run build
echo "✅ Mobile build success (mobile/dist)"
cd ..

echo "==========================================="
echo "🎉 All builds completed successfully!"
echo "==========================================="
