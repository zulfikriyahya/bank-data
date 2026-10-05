#!/bin/bash
set -e

cd "$(dirname "$0")"
source venv/bin/activate

COOKIES_FILE="${1:-cookies_export.json}"
LOCALSTORAGE_FILE="${2:-localstorage_export.json}"
MODE="${3:-resume}"

if [ ! -f "$COOKIES_FILE" ]; then
    echo "[ERROR] File cookies tidak ditemukan: $COOKIES_FILE"
    echo "Usage: ./import_from_cookie.sh <cookies.json> <localstorage.json> [mode]"
    exit 1
fi

if [ ! -f "$LOCALSTORAGE_FILE" ]; then
    echo "[ERROR] File localStorage tidak ditemukan: $LOCALSTORAGE_FILE"
    exit 1
fi

echo "=== 1. Menggabungkan cookie + localStorage jadi auth.json ==="
python3 merge_auth.py "$COOKIES_FILE" "$LOCALSTORAGE_FILE" auth.json

echo ""
echo "=== 2. Validasi sesi masih aktif ==="
VALID=$(python3 -c "
from playwright.sync_api import sync_playwright
with sync_playwright() as p:
    browser = p.firefox.launch(headless=True)
    context = browser.new_context(storage_state='auth.json')
    page = context.new_page()
    try:
        page.goto('https://emis.kemenag.go.id/kesiswaan?page=1', wait_until='networkidle', timeout=30000)
        count = page.locator('text=Aksi').count()
        print(count)
    except Exception as e:
        print(0)
    browser.close()
")

if [ "$VALID" -eq 0 ]; then
    echo "[ERROR] Sesi tidak valid atau EMIS tidak merespon. Scraping dibatalkan."
    echo "Pastikan kamu baru saja login dan export cookie+localStorage dalam kondisi masih aktif."
    exit 1
fi

echo "Sesi valid, ditemukan $VALID baris di halaman 1."
echo ""
echo "=== 3. Menjalankan scraping (mode: $MODE) ==="
python3 scraper.py "$MODE"

echo ""
echo "=== 4. Bersihkan file export mentah ==="
rm -f "$COOKIES_FILE" "$LOCALSTORAGE_FILE"

echo ""
echo "=== Selesai ==="
