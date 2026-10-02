## Setup
sudo apt install python3-pip python3-venv -y
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
python3 -m playwright install --with-deps firefox

## Pemakaian
python3 scraper.py          # scraping dari awal
python3 scraper.py resume   # lanjutkan progress yang terhenti
python3 scraper.py retry    # ulangi halaman yang belum lengkap
