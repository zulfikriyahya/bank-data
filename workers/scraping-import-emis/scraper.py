import os
import sys
import json
import atexit
import io
import threading
from pathlib import Path
from collections import Counter
from concurrent.futures import ThreadPoolExecutor, as_completed
from dotenv import load_dotenv
from playwright.sync_api import sync_playwright
from bs4 import BeautifulSoup
from tqdm import tqdm
import pandas as pd
import requests
from datetime import datetime

load_dotenv()

class _FilteredStderr(io.TextIOBase):
    """Menyaring noise asyncio/Playwright saat cleanup interpreter
    (Task destroyed / TargetClosedError / Future exception never retrieved).
    Pesan-pesan ini muncul setelah proses utama selesai sukses dan
    tidak berpengaruh pada hasil scraping."""

    _NOISE_MARKERS = (
        "Task was destroyed but it is pending",
        "TargetClosedError",
        "Future exception was never retrieved",
        "coro=<Connection.run",
        "Target page, context or browser has been closed",
    )

    def __init__(self, original):
        self._original = original
        self._buffer = ""

    def write(self, s):
        self._buffer += s
        while "\n" in self._buffer:
            line, self._buffer = self._buffer.split("\n", 1)
            if not any(marker in line for marker in self._NOISE_MARKERS):
                self._original.write(line + "\n")
        return len(s)

    def flush(self):
        if self._buffer and not any(m in self._buffer for m in self._NOISE_MARKERS):
            self._original.write(self._buffer)
        self._buffer = ""
        self._original.flush()


def _install_stderr_filter():
    sys.stderr = _FilteredStderr(sys.stderr)
    atexit.register(lambda: sys.stderr.flush())


_install_stderr_filter()


def fetch_config_from_core_app():
    """Ambil kredensial & config dari core-app, bukan dari .env lokal lagi.
    .env lokal cuma simpan CORE_APP_URL dan SERVICE_TOKEN_EMIS sebagai kunci akses."""
    core_app_url = os.getenv("CORE_APP_URL", "http://localhost:8080/api/v1/internal")
    service_token = os.getenv("SERVICE_TOKEN_EMIS")

    if not service_token:
        print("[CONFIG ERROR] SERVICE_TOKEN_EMIS belum diisi di .env")
        sys.exit(1)

    url = f"{core_app_url}/scraper-config/emis-scraper"
    headers = {"Authorization": f"Bearer {service_token}"}

    try:
        response = requests.get(url, headers=headers, timeout=10)
        if response.status_code != 200:
            print(f"[CONFIG ERROR] Gagal ambil config dari core-app: HTTP {response.status_code} - {response.text}")
            sys.exit(1)
        return response.json().get("config", {})
    except requests.exceptions.RequestException as e:
        print(f"[CONFIG ERROR] Tidak bisa konek ke core-app untuk ambil config: {e}")
        sys.exit(1)


_remote_config = fetch_config_from_core_app()

EMAIL = _remote_config.get("EMIS_EMAIL")
PASSWORD = _remote_config.get("EMIS_PASSWORD")
BASE_URL = _remote_config.get("BASE_URL", "https://emis.kemenag.go.id/kesiswaan")
ROWS_PER_PAGE = int(_remote_config.get("ROWS_PER_PAGE", "30"))
WORKERS = int(_remote_config.get("SCRAPER_WORKERS", "1"))

AUTH_FILE = os.getenv("AUTH_FILE", "auth.json")
BACKUP_FILE = os.getenv("BACKUP_FILE", "progress_backup.json")
OUTPUT_FILE = os.getenv("OUTPUT_FILE", "data_siswa_emis.xlsx")

TOTAL_PAGES_OVERRIDE = os.getenv("TOTAL_PAGES_OVERRIDE")
LAST_PAGE_COUNT_OVERRIDE = os.getenv("LAST_PAGE_COUNT_OVERRIDE")

NAV_TIMEOUT_MS = int(os.getenv("NAV_TIMEOUT_MS", "45000"))
SELECTOR_TIMEOUT_MS = int(os.getenv("SELECTOR_TIMEOUT_MS", "20000"))
DETAIL_WAIT_MS = int(os.getenv("DETAIL_WAIT_MS", "1000"))
RETRY_WAIT_MS = int(os.getenv("RETRY_WAIT_MS", "1500"))
POST_BACK_WAIT_MS = int(os.getenv("POST_BACK_WAIT_MS", "500"))
PROBE_TIMEOUT_MS = int(os.getenv("PROBE_TIMEOUT_MS", "10000"))
PROBE_RETRIES = int(os.getenv("PROBE_RETRIES", "1"))
WARMUP_TIMEOUT_MS = int(os.getenv("WARMUP_TIMEOUT_MS", "60000"))

CORE_APP_URL = os.getenv("CORE_APP_URL", "http://localhost:8080/api/v1/internal")
SERVICE_TOKEN_EMIS = os.getenv("SERVICE_TOKEN_EMIS")
IMPORT_MODE = os.getenv("IMPORT_MODE", "upsert")
PUSH_TO_CORE_APP = os.getenv("PUSH_TO_CORE_APP", "true").lower() == "true"
PUSH_FAILURES_FILE = os.getenv("PUSH_FAILURES_FILE", "push_failures.json")
PUSH_TIMEOUT_SEC = int(os.getenv("PUSH_TIMEOUT_SEC", "30"))

_push_lock = threading.Lock()

TABLE_SECTION_TITLES = {
    "AKTIVITAS BELAJAR": "aktivitas_belajar",
    "BEASISWA & BANTUAN": "beasiswa",
    "PRESTASI SISWA": "prestasi",
}
PERSON_SECTION_KEYS = {
    "Ayah kandung": "ayah",
    "Ibu kandung": "ibu",
    "Wali": "wali",
}


# ---------- validasi awal ----------

def validate_config():
    errors = []
    if not EMAIL:
        errors.append("EMIS_EMAIL belum diisi di .env")
    if not PASSWORD:
        errors.append("EMIS_PASSWORD belum diisi di .env")
    if WORKERS < 1:
        errors.append("SCRAPER_WORKERS harus >= 1")
    if PUSH_TO_CORE_APP and not SERVICE_TOKEN_EMIS:
        errors.append("SERVICE_TOKEN_EMIS belum diisi di .env (atau set PUSH_TO_CORE_APP=false)")
    if errors:
        for e in errors:
            print(f"[CONFIG ERROR] {e}")
        sys.exit(1)


# ---------- login ----------

def ensure_login():
    """Login manual sekali (captcha), simpan session ke AUTH_FILE.
    Kalau tidak ada auth.json dan berjalan di mode non-interaktif
    (misal dari cron/PM2), gagal dengan jelas, jangan hang nunggu GUI."""
    if Path(AUTH_FILE).exists():
        return

    is_interactive = sys.stdin.isatty()
    if not is_interactive:
        print("[FATAL] auth.json tidak ditemukan dan proses berjalan non-interaktif (cron/PM2).")
        print("[FATAL] Generate auth.json secara manual dulu di mesin dengan GUI, lalu upload ke server.")
        sys.exit(1)

    with sync_playwright() as playwright:
        browser = playwright.firefox.launch(headless=False)
        context = browser.new_context()
        page = context.new_page()
        page.goto("https://emis.kemenag.go.id/")
        page.get_by_role("button", name="Login").click()
        page.locator('input[type="email"]').fill(EMAIL)
        page.locator('input[type="password"]').fill(PASSWORD)
        print("\n>>> Selesaikan captcha 'Saya bukan robot' secara manual di window Firefox.")
        print(">>> Setelah berhasil login dan halaman utama EMIS muncul, kembali ke terminal ini dan tekan ENTER.\n")
        input("Tekan ENTER setelah login manual selesai...")
        context.storage_state(path=AUTH_FILE)
        browser.close()


def new_context(playwright, headless=True):
    browser = playwright.firefox.launch(headless=headless)
    context = browser.new_context(storage_state=AUTH_FILE)
    return browser, context


def warmup_context(page, pbar=None):
    """Goto pertama pada context baru cenderung lambat (cold start).
    Buang goto pertama ke halaman home dulu supaya goto berikutnya stabil."""
    try:
        page.goto("https://emis.kemenag.go.id/", timeout=WARMUP_TIMEOUT_MS, wait_until="load")
    except Exception as e:
        msg = f"[WARN] Warm-up gagal (lanjut tetap): {e}"
        if pbar:
            pbar.write(msg)
        else:
            print(msg)


# ---------- pagination probe ----------

def count_rows_on_page(page, page_num, retries=PROBE_RETRIES) -> int:
    for attempt in range(retries + 1):
        try:
            page.goto(f"{BASE_URL}?page={page_num}", wait_until="networkidle", timeout=NAV_TIMEOUT_MS)
            page.wait_for_selector("text=Aksi", timeout=PROBE_TIMEOUT_MS)
            rows = page.locator('div.r-14lw9ot[style*="border-radius: 0px"]').filter(has_text="Aksi")
            return rows.count()
        except Exception as e:
            if attempt < retries:
                print(f"[WARN] page {page_num} gagal (percobaan {attempt + 1}), retry: {e}")
                page.wait_for_timeout(2000)
                continue
            print(f"[WARN] page {page_num} dianggap kosong setelah {retries + 1}x percobaan: {e}")
            return 0
    return 0


def detect_total_pages(playwright) -> int:
    """Cari halaman terakhir yang punya data, dengan galloping search lalu bisection."""
    browser, context = new_context(playwright)
    page = context.new_page()
    try:
        warmup_context(page)
        print("Mendeteksi jumlah halaman total...")
        lo, hi = 1, 1
        if count_rows_on_page(page, 1) == 0:
            print("[WARN] Page 1 kosong atau gagal dimuat — cek auth.json atau koneksi.")
            return 0

        while count_rows_on_page(page, hi) > 0:
            lo = hi
            hi *= 2
            if hi > 2000:
                break

        while lo < hi - 1:
            mid = (lo + hi) // 2
            if count_rows_on_page(page, mid) > 0:
                lo = mid
            else:
                hi = mid

        print(f"Total halaman terdeteksi: {lo}")
        return lo
    finally:
        browser.close()


def get_total_pages(playwright) -> int:
    if TOTAL_PAGES_OVERRIDE:
        total = int(TOTAL_PAGES_OVERRIDE)
        print(f"Menggunakan total halaman manual (dari .env): {total}")
        return total
    return detect_total_pages(playwright)


def get_last_page_count(playwright, total_pages: int) -> int:
    if LAST_PAGE_COUNT_OVERRIDE:
        count = int(LAST_PAGE_COUNT_OVERRIDE)
        print(f"Menggunakan jumlah baris halaman terakhir manual (dari .env): {count}")
        return count
    browser, context = new_context(playwright)
    p = context.new_page()
    warmup_context(p)
    count = count_rows_on_page(p, total_pages)
    browser.close()
    return count


# ---------- parsing ----------

def parse_person_section(section_tag):
    data = {}
    rows = section_tag.find_all(
        "div",
        class_=lambda c: c and all(
            k in c.split() for k in ["r-18u37iz", "r-1wtj0ep", "r-oyd9sg", "r-13qz1uu"]
        ),
        recursive=False,
    )
    for row in rows:
        if row.select_one('[data-testid="text-name"]'):
            continue
        label_el = row.select_one('div[class*="r-1enofrn"]')
        value_el = row.select_one('[data-testid="text-address-value"]')
        if label_el and value_el:
            data[label_el.get_text(strip=True)] = value_el.get_text(strip=True)
    return data


def parse_table_section(section_tag):
    if section_tag.select_one('[data-testid="no-data"]'):
        return []
    header_row = section_tag.select_one('div[style*="padding: 10px"] > div')
    if not header_row:
        return []
    headers = [c.get_text(strip=True) for c in header_row.find_all("div", recursive=False)]
    rows_container = section_tag.select_one(".r-tvv088")
    if not rows_container:
        return []
    records = []
    for row in rows_container.find_all("div", recursive=False):
        cells = [c.get_text(strip=True) for c in row.find_all("div", recursive=False)]
        if len(cells) == len(headers):
            records.append(dict(zip(headers, cells)))
    return records


def parse_detail_html(html: str) -> dict:
    soup = BeautifulSoup(html, "html.parser")
    root = soup.select_one('div[style*="padding-bottom: 20px"]') or soup

    result = {"siswa": {}, "ayah": {}, "ibu": {}, "wali": {},
              "aktivitas_belajar": [], "beasiswa": [], "prestasi": []}

    headers = root.select('[data-testid="text-name"]')
    for header in headers:
        title_el = header.select_one('[data-testid="text-name-value"]')
        title = title_el.get_text(strip=True) if title_el else ""
        section_container = header.parent.parent

        if title in TABLE_SECTION_TITLES:
            result[TABLE_SECTION_TITLES[title]] = parse_table_section(section_container)
        elif title in PERSON_SECTION_KEYS:
            result[PERSON_SECTION_KEYS[title]] = parse_person_section(section_container)
        else:
            result["siswa"] = parse_person_section(section_container)
            result["siswa"]["NAMA"] = title

    return result


def is_record_valid(record) -> bool:
    siswa = record.get("siswa", {})
    return bool(siswa.get("NIK") or siswa.get("NAMA") or siswa.get("NISN"))


# ---------- per-page scraping ----------

def scrape_page(page, page_num, out_records, pbar=None):
    try:
        page.goto(f"{BASE_URL}?page={page_num}", wait_until="networkidle", timeout=NAV_TIMEOUT_MS)
        page.wait_for_selector("text=Aksi", timeout=SELECTOR_TIMEOUT_MS)
    except Exception as e:
        if pbar:
            pbar.write(f"Gagal buka page {page_num}: {e}")
        return

    rows = page.locator('div.r-14lw9ot[style*="border-radius: 0px"]').filter(has_text="Aksi")
    row_count = rows.count()

    for i in range(row_count):
        if pbar:
            pbar.set_description(f"Page {page_num} | Baris {i + 1}/{row_count}")
        try:
            row = rows.nth(i)
            row.scroll_into_view_if_needed(timeout=45000)
            row.get_by_text("Aksi", exact=True).click()
            page.get_by_role("menuitem", name="Lihat Detail").click()

            page.wait_for_selector('[data-testid="text-name-value"]', state="visible", timeout=SELECTOR_TIMEOUT_MS)
            page.wait_for_load_state("networkidle")
            page.wait_for_timeout(DETAIL_WAIT_MS)
            page.wait_for_function(
                """() => {
                    const headers = document.querySelectorAll('[data-testid="text-name-value"]');
                    return Array.from(headers).some(h => h.textContent.includes('PRESTASI SISWA'));
                }""",
                timeout=10000,
            )

            heading = page.locator('[data-testid="text-name"]').first
            detail_container = heading.locator(
                'xpath=ancestor::div[@style and contains(@style, "padding-bottom: 20px")]'
            ).first
            detail_container.wait_for(state="visible", timeout=10000)
            html = detail_container.evaluate("el => el.outerHTML")
            record = parse_detail_html(html)

            if not is_record_valid(record):
                if pbar:
                    pbar.write(f"Data siswa kosong di page {page_num} baris {i + 1}, retry...")
                page.wait_for_timeout(RETRY_WAIT_MS)
                html = detail_container.evaluate("el => el.outerHTML")
                record = parse_detail_html(html)

            record["_page"] = page_num
            record["_row"] = i + 1
            record["_failed"] = not is_record_valid(record)
            out_records.append(record)

            if pbar:
                pbar.write(f"Page {page_num} baris {i + 1}: {record['siswa'].get('NAMA', '-')}")

            page.go_back()
            page.wait_for_load_state("networkidle")
            page.wait_for_timeout(POST_BACK_WAIT_MS)

        except Exception as e:
            if pbar:
                pbar.write(f"Gagal di page {page_num} baris {i + 1}: {e}")
            try:
                page.screenshot(path=f"debug_page{page_num}_row{i + 1}.png", full_page=True)
            except Exception:
                pass
            out_records.append({
                "siswa": {}, "ayah": {}, "ibu": {}, "wali": {},
                "aktivitas_belajar": [], "beasiswa": [], "prestasi": [],
                "_page": page_num, "_row": i + 1, "_failed": True,
            })
            try:
                page.goto(f"{BASE_URL}?page={page_num}", wait_until="networkidle", timeout=NAV_TIMEOUT_MS)
            except Exception:
                pass
            continue
        finally:
            if pbar:
                pbar.update(1)


# ---------- worker (own playwright instance per thread) ----------

def worker(page_numbers, shared_records, shared_lock, pbar):
    with sync_playwright() as playwright:
        browser, context = new_context(playwright)
        page = context.new_page()
        try:
            warmup_context(page, pbar)
            for page_num in page_numbers:
                local = []
                try:
                    scrape_page(page, page_num, local, pbar)
                except Exception as e:
                    pbar.write(f"Gagal total di page {page_num}: {e}")
                with shared_lock:
                    shared_records.extend(local)
                    with open(BACKUP_FILE, "w", encoding="utf-8") as f:
                        json.dump(shared_records, f, ensure_ascii=False, indent=2)

                # baris baru ini yang ditambahkan:
                push_page_to_core_app(page_num, local, pbar)
        finally:
            browser.close()


def run_pages_parallel(page_list, existing_records):
    all_records = list(existing_records)
    lock = threading.Lock()

    buckets = [page_list[i::WORKERS] for i in range(WORKERS)]
    buckets = [b for b in buckets if b]

    total_estimate = len(page_list) * ROWS_PER_PAGE
    with tqdm(total=total_estimate, unit="siswa") as pbar:
        with ThreadPoolExecutor(max_workers=len(buckets)) as executor:
            futures = [executor.submit(worker, b, all_records, lock, pbar) for b in buckets]
            for f in as_completed(futures):
                f.result()

    return all_records


# ---------- excel export / dup check ----------

def check_duplicate_nisn(records):
    nisn_list = [r["siswa"].get("NISN", "") for r in records if r["siswa"].get("NISN")]
    counter = Counter(nisn_list)
    duplicates = {nisn: count for nisn, count in counter.items() if count > 1}
    if duplicates:
        print(f"\nDitemukan {len(duplicates)} NISN duplikat:")
        for nisn, count in duplicates.items():
            print(f"  - {nisn}: muncul {count}x")
    else:
        print("\nTidak ada NISN duplikat.")
    return duplicates

def build_combined_rows(valid_records):
    """Gabungkan data siswa + ayah + ibu + wali jadi satu baris per siswa,
    dengan prefix kolom supaya tidak bentrok antar kategori."""
    combined_rows = []
    for r in valid_records:
        row = {"nama": r["siswa"].get("NAMA", ""), **r["siswa"]}

        for label, key in (("ayah", "ayah"), ("ibu", "ibu"), ("wali", "wali")):
            for col_name, val in r.get(key, {}).items():
                row[f"{label}_{col_name}"] = val

        combined_rows.append(row)
    return combined_rows

def export_to_excel(records, filename=OUTPUT_FILE):
    check_duplicate_nisn(records)
    valid_records = [r for r in records if not r.get("_failed") and is_record_valid(r)]
    skipped = len(records) - len(valid_records)
    if skipped:
        print(f"{skipped} record gagal/kosong tidak diikutkan ke Excel.")

    siswa_rows, ayah_rows, ibu_rows, wali_rows = [], [], [], []
    aktivitas_rows, beasiswa_rows, prestasi_rows = [], [], []

    for r in valid_records:
        nisn = r["siswa"].get("NISN", "")
        nama = r["siswa"].get("NAMA", "")
        siswa_rows.append({"nama": nama, **r["siswa"]})
        if r["ayah"]:
            ayah_rows.append({"nisn": nisn, "nama_siswa": nama, **r["ayah"]})
        if r["ibu"]:
            ibu_rows.append({"nisn": nisn, "nama_siswa": nama, **r["ibu"]})
        if r["wali"]:
            wali_rows.append({"nisn": nisn, "nama_siswa": nama, **r["wali"]})
        for row in r["aktivitas_belajar"]:
            aktivitas_rows.append({"nisn": nisn, "nama_siswa": nama, **row})
        for row in r["beasiswa"]:
            beasiswa_rows.append({"nisn": nisn, "nama_siswa": nama, **row})
        for row in r["prestasi"]:
            prestasi_rows.append({"nisn": nisn, "nama_siswa": nama, **row})

    combined_rows = build_combined_rows(valid_records)

    with pd.ExcelWriter(filename, engine="openpyxl") as writer:
        pd.DataFrame(combined_rows).to_excel(writer, sheet_name="Semua Data", index=False)
        pd.DataFrame(siswa_rows).to_excel(writer, sheet_name="Siswa", index=False)
        pd.DataFrame(ayah_rows).to_excel(writer, sheet_name="Ayah", index=False)
        pd.DataFrame(ibu_rows).to_excel(writer, sheet_name="Ibu", index=False)
        pd.DataFrame(wali_rows).to_excel(writer, sheet_name="Wali", index=False)
        pd.DataFrame(aktivitas_rows).to_excel(writer, sheet_name="Aktivitas Belajar", index=False)
        pd.DataFrame(beasiswa_rows).to_excel(writer, sheet_name="Beasiswa", index=False)
        pd.DataFrame(prestasi_rows).to_excel(writer, sheet_name="Prestasi", index=False)

    print(f"Selesai. Data tersimpan di {filename}")


def load_existing_records():
    if Path(BACKUP_FILE).exists():
        with open(BACKUP_FILE, "r", encoding="utf-8") as f:
            return json.load(f)
    return []

# ---------- push ke core-app ----------

def log_push_failure(page_num, records_count, error_message, pbar=None):
    entry = {
        "page": page_num,
        "records_count": records_count,
        "error": str(error_message),
        "timestamp": datetime.now().isoformat(),
    }

    with _push_lock:
        failures = []
        if Path(PUSH_FAILURES_FILE).exists():
            try:
                with open(PUSH_FAILURES_FILE, "r", encoding="utf-8") as f:
                    failures = json.load(f)
            except (json.JSONDecodeError, FileNotFoundError):
                failures = []

        failures.append(entry)

        with open(PUSH_FAILURES_FILE, "w", encoding="utf-8") as f:
            json.dump(failures, f, ensure_ascii=False, indent=2)

    msg = f"[PUSH FAILED] Page {page_num}: {error_message} (dicatat ke {PUSH_FAILURES_FILE})"
    if pbar:
        pbar.write(msg)
    else:
        print(msg)


def push_page_to_core_app(page_num, records, pbar=None):
    if not PUSH_TO_CORE_APP:
        return

    valid_records = [r for r in records if not r.get("_failed") and is_record_valid(r)]

    if not valid_records:
        if pbar:
            pbar.write(f"Page {page_num}: tidak ada record valid untuk di-push, dilewati.")
        return

    url = f"{CORE_APP_URL}/import/emis?mode={IMPORT_MODE}"
    headers = {
        "Authorization": f"Bearer {SERVICE_TOKEN_EMIS}",
        "Content-Type": "application/json",
    }

    try:
        response = requests.post(url, json=valid_records, headers=headers, timeout=PUSH_TIMEOUT_SEC)

        if response.status_code in (200, 201, 202):
            result = response.json()
            summary = result.get("result", {})
            msg = (f"Page {page_num} ter-push: "
                   f"{summary.get('success_rows', '?')} sukses, "
                   f"{summary.get('failed_rows', '?')} gagal, "
                   f"{summary.get('duplicate_rows', '?')} duplikat")
            if pbar:
                pbar.write(msg)
            else:
                print(msg)
        else:
            log_push_failure(page_num, len(valid_records), f"HTTP {response.status_code}: {response.text[:200]}", pbar)

    except requests.exceptions.RequestException as e:
        log_push_failure(page_num, len(valid_records), str(e), pbar)

# ---------- mode logic ----------

def pages_with_full_count(records, total_pages, expected_last_page_count):
    counted = {}
    failed_pages = set()
    for r in records:
        p = r["_page"]
        counted[p] = counted.get(p, 0) + 1
        if r.get("_failed") or not is_record_valid(r):
            failed_pages.add(p)

    complete = set()
    for p, c in counted.items():
        expected = expected_last_page_count if p == total_pages else ROWS_PER_PAGE
        if c >= expected and p not in failed_pages:
            complete.add(p)
    return complete, failed_pages


def compute_pages_to_run(existing, total_pages, last_page_count):
    """Dipakai oleh mode retry & resume — keduanya sama-sama
    membandingkan terhadap rentang penuh 1..total_pages, bukan
    cuma halaman yang kebetulan sudah punya record (halaman yang
    gagal total / goto error tidak akan muncul di 'existing')."""
    complete, _failed = pages_with_full_count(existing, total_pages, last_page_count)
    pages_to_run = sorted(set(range(1, total_pages + 1)) - complete)
    existing_filtered = [r for r in existing if r["_page"] not in pages_to_run]
    return pages_to_run, existing_filtered, complete


def main():
    valid_modes = {"retry", "resume"}
    mode = sys.argv[1] if len(sys.argv) > 1 else None
    if mode and mode not in valid_modes:
        print(f"Mode tidak dikenal: '{mode}'. Gunakan salah satu: (kosong)/retry/resume.")
        sys.exit(1)

    retry_mode = mode == "retry"
    resume_mode = mode == "resume"

    validate_config()
    ensure_login()

    with sync_playwright() as playwright:
        total_pages = get_total_pages(playwright)
        if total_pages == 0:
            print("[FATAL] Tidak ada data ditemukan — kemungkinan auth.json sudah expired.")
            print("[FATAL] Generate ulang auth.json dari mesin dengan GUI.")
            sys.exit(1)
            return
        last_page_count = get_last_page_count(playwright, total_pages)

    existing = load_existing_records() if (retry_mode or resume_mode) else []

    if retry_mode or resume_mode:
        pages_to_run, existing, complete = compute_pages_to_run(existing, total_pages, last_page_count)
        label = "di-retry" if retry_mode else "diproses (resume)"
        print(f"Halaman selesai: {len(complete)}/{total_pages}. Page yang akan {label}: {pages_to_run}")
    else:
        pages_to_run = list(range(1, total_pages + 1))
        existing = []

    if not pages_to_run:
        print("Tidak ada halaman yang perlu diproses.")
        all_records = existing
    else:
        all_records = run_pages_parallel(pages_to_run, existing)

    export_to_excel(all_records)


if __name__ == "__main__":
    main()
