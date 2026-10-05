#!/bin/bash
set -e

cd "$(dirname "$0")"
source venv/bin/activate

MODE="${1:-resume}"

echo "=== Validasi sesi ==="
VALID=$(python3 -c "
from playwright.sync_api import sync_playwright
with sync_playwright() as p:
    browser = p.firefox.launch(headless=True)
    context = browser.new_context(storage_state='auth.json')
    page = context.new_page()
    try:
        page.goto('https://emis.kemenag.go.id/kesiswaan?page=1', wait_until='networkidle', timeout=30000)
        print(page.locator('text=Aksi').count())
    except Exception:
        print(0)
    browser.close()
")

if [ "$VALID" -eq 0 ]; then
    echo "[ERROR] Sesi tidak valid. Pastikan masih login saat extension dijalankan."
    exit 1
fi

echo "Sesi valid, $VALID baris ditemukan di halaman 1."
echo "=== Menjalankan scraping (mode: $MODE) ==="
python3 scraper.py "$MODE"
echo "=== Selesai ==="
