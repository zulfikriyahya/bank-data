#!/bin/bash
cd "$(dirname "$0")"
source venv/bin/activate
python3 scraper.py
read -p "Selesai. Tekan ENTER untuk menutup..."
# di akhir run_scraper.sh atau cron job, cek exit code
python3 scraper.py resume
if [ $? -ne 0 ]; then
    echo "Scraping EMIS gagal pada $(date)" | mail -s "ALERT: Scraper EMIS Gagal" admin@email.com
    # atau kirim ke Slack/Telegram webhook kalau sudah ada infrastrukturnya
fi
