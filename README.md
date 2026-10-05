# SOFTWARE REQUIREMENTS SPECIFICATION (SRS)
# Sistem Bank Data — Master Data Management Siswa Madrasah
**MTsN 1 Pandeglang**

Versi: 1.0
Tanggal: 5 Oktober 2026
Status: Fase inti (core-app + scraping-import-emis + extension) **selesai & terverifikasi**. Dokumen ini disusun untuk melanjutkan pengerjaan oleh model AI / engineer berikutnya.

---

## 1. Pendahuluan

### 1.1 Tujuan Dokumen
Dokumen ini adalah spesifikasi final dan terstruktur dari sistem "Bank Data" — sebuah platform Master Data Management (MDM) untuk data identitas siswa madrasah — mencakup apa yang sudah dibangun, keputusan arsitektur yang diambil (dan alasannya), bug kritis yang pernah terjadi beserta perbaikannya, serta roadmap tahap berikutnya. Dokumen ini ditulis secara eksplisit agar dapat digunakan untuk melatih/mengorientasikan model AI lain yang akan melanjutkan proyek ini tanpa kehilangan konteks.

### 1.2 Lingkup
Sistem ini adalah **Single Source of Truth (SSOT)** untuk data biodata siswa (dan data keluarga terkait) yang dikonsumsi oleh beberapa aplikasi madrasah lain (presensi, perpustakaan), dan disinkronkan dari portal EMIS Kementerian Agama.

### 1.3 Definisi & Istilah
Lihat Glosarium (Bagian 12).

---

## 2. Gambaran Umum Sistem

### 2.1 Latar Belakang
Madrasah memiliki beberapa aplikasi yang masing-masing menyimpan data siswa secara terpisah dan tidak sinkron (presensi berbasis PHP Laravel, perpustakaan berbasis PHP Laravel), serta harus menyesuaikan data dengan portal EMIS milik Kementerian Agama yang **tidak memiliki API resmi** (hanya bisa diakses lewat scraping HTML dengan login manual + captcha).

### 2.2 Tujuan Sistem
1. Menjadi satu sumber data (SSOT) biodata siswa + keluarga + aktivitas belajar + beasiswa + prestasi.
2. Menyediakan API terpusat agar aplikasi presensi dan perpustakaan tidak lagi menyimpan biodata siswa sendiri-sendiri, cukup konsumsi dari core-app.
3. Mengotomasi penarikan data dari EMIS (scraping) dan mendorongnya (push) ke core-app tanpa campur tangan manual di server, termasuk tanpa akses GUI/SSH langsung ke server produksi.
4. Menyimpan kredensial sensitif (kredensial EMIS) secara terenkripsi, bukan sebagai plaintext di file `.env`.

### 2.3 Arsitektur Tingkat Tinggi
```
                         ┌─────────────────────┐
   (scraping-import-emis)│                      │  (import-data-pendaftaran-app)
   Browser Extension ───▶│                      │◀───  (aplikasi pendaftaran, existing)
   (capture session)      │      core-app        │
                         │  (Go + Fiber + MySQL)│
   EMIS Kemenag ◀── scrape│   SSOT biodata siswa │
   (no official API)      │                      │
                         └─────────┬────────────┘
                                   │ (consumer API, scoped API key)
                      ┌────────────┼─────────────┐
                      ▼                          ▼
              app-presensi                app-perpustakaan
          (PHP Laravel, existing)      (PHP CodeIgniter, existing)
             — read siswa data,            — read siswa data,
             — write lightweight            — write lightweight
               activity_log only              activity_log only
```

Keputusan kunci: **tidak menduplikasi data bisnis presensi/perpustakaan ke core-app.** Core-app hanya menyimpan jejak aktivitas ringan (`activity_log`), bukan data detail presensi/peminjaman buku. (Lihat ADR-02, Bagian 13.)

### 2.4 Batasan Teknis yang Mengikat
- **Tidak menggunakan Docker.** Deployment memakai **PM2** (proses Node.js process manager) untuk mengelola binary Go dan worker Python. Ini adalah keputusan eksplisit dari pemilik proyek, bukan default teknis — jangan diubah tanpa persetujuan ulang.
- Server produksi **tidak memiliki GUI** — semua proses interaktif (login EMIS dengan captcha) harus diselesaikan di luar server lalu hasil sesi dikirim ke server.
- EMIS **tidak memiliki API resmi** — satu-satunya cara mengambil data adalah scraping HTML halaman kesiswaan setelah login.

---

## 3. Struktur Data & Skema Database

### 3.1 Skema JSON Kanonikal (sumber kebenaran struktur data)
Field inti per siswa, dipecah ke beberapa tabel relasional:
```json
{
  "siswa": {
    "nik": "", "nisn": "", "kip": "", "nama": "", "tempat_lahir": "",
    "tanggal_lahir": "", "jenis_kelamin": "", "agama": "", "jumlah_saudara": 0,
    "anak_ke": 0, "hobi": "", "cita_cita": "", "no_handphone": "", "email": "",
    "yang_membiayai": "", "kebutuhan_disabilitas": "", "kebutuhan_khusus": "",
    "alamat": "", "status_tempat_tinggal": "", "jarak_rumah_madrasah": "",
    "waktu_tempuh": "", "transportasi": "", "fingerprint": "", "rfid": ""
  },
  "ayah": { "...biodata ayah..." },
  "ibu": { "...biodata ibu..." },
  "wali": { "...biodata wali, opsional..." },
  "aktivitas_belajar": [ { "...riwayat kelas/tahun ajaran..." } ],
  "beasiswa": [ { "...riwayat beasiswa..." } ],
  "prestasi": [ { "...riwayat prestasi..." } ],
  "_page": 0, "_row": 0, "_failed": false
}
```
`_page`, `_row`, `_failed` adalah metadata hasil scraping (bukan bagian biodata), dipetakan ke kolom `source_page`, `source_row`, `is_failed` pada tabel `siswa`.

### 3.2 Daftar Tabel (10 migrasi, berurutan & reversible)
1. `siswa` — tabel inti biodata siswa (termasuk `uuid`, `source` enum: `pendaftaran`/`emis_scraping`/manual, `source_page`, `source_row`, `is_failed`, timestamps).
2. `wali_siswa` — menyatukan ayah/ibu/wali dengan kolom `tipe` (ayah/ibu/wali) relasi ke `siswa.id`.
3. `aktivitas_belajar` — riwayat kelas per tahun ajaran, relasi ke `siswa.id`.
4. `beasiswa` dan `prestasi` — riwayat terpisah, relasi ke `siswa.id`.
5. `import_log` + `import_log_detail` — jejak setiap proses import (batch), status per baris.
6. `api_clients` + `api_client_access_log` — klien konsumen (presensi/perpustakaan) dengan API key (bcrypt hash) + scope JSON array + log akses.
7. `admin_users` — akun admin (bcrypt password) untuk JWT auth.
8. `activity_log` — log ringan aktivitas dari aplikasi konsumen (lihat ADR-02).
9. `scraper_configs` — konfigurasi worker scraping, termasuk kredensial EMIS terenkripsi AES-GCM.

### 3.3 Aturan Normalisasi
- Tanggal Indonesia (`"21 April 2014"`) dinormalisasi ke format SQL `DATE` via `pkg/normalizer.ParseTanggalIndonesia`.
- String kosong/`"-"`/`"N/A"` dinormalisasi ke SQL `NULL` via `NullableString`/`NullableInt`.
- `jenis_kelamin` dinormalisasi ke `"Laki-laki"`/`"Perempuan"` baku via `NormalizeJenisKelamin`.

---

## 4. Spesifikasi Modul: Core App (Go + Fiber + MySQL)

### 4.1 Arsitektur Kode
Layered: `handler → service → repository → domain`, dengan `pkg/` untuk utilitas lintas-layer (jsonparser, normalizer, excelgen, cryptoutil, jwtutil).

### 4.2 Endpoint — Ringkasan per Grup

**Grup `internal/*` (auth: static bearer Service Token, dipakai worker↔core-app)**
| Endpoint | Metode | Fungsi |
|---|---|---|
| `/internal/import` | POST | Import JSON batch (insert/upsert) dari worker pendaftaran/scraper |
| `/internal/scraper-config/:name` | GET | Ambil konfigurasi scraper (kredensial didekripsi) |

**Grup `consumer/*` (auth: API Key berskop, dipakai app-presensi/app-perpustakaan)**
| Endpoint | Metode | Fungsi | Scope dibutuhkan |
|---|---|---|---|
| `/consumer/siswa/:id` | GET | Detail siswa (ringkas) | `siswa:read` |
| `/consumer/siswa/search?rfid=/nisn=/nik=` | GET | Cari satu siswa by field (whitelist field) | `siswa:read` |
| `/consumer/activity-log` | POST | Catat jejak aktivitas ringan | `activity:write` |

**Grup `admin/*` (auth: JWT admin)**
| Endpoint | Metode | Fungsi |
|---|---|---|
| `/auth/login` | POST | Login admin, keluarkan JWT |
| `/admin/siswa` | GET/POST/PUT/DELETE | CRUD penuh data siswa |
| `/admin/siswa/export?format=json\|xlsx` | GET | Export data |
| `/admin/import` | POST | Upload file JSON untuk import manual |
| `/admin/api-clients` | GET/POST/PUT/DELETE | CRUD klien konsumen + regenerate key |
| `/admin/activity-log` | GET | Lihat jejak aktivitas aplikasi konsumen |
| `/admin/scraper-config/:name` | GET/PUT | Lihat (masked)/ubah konfigurasi scraper |
| `/admin/scraper/emis/trigger-json` | POST | **(baru)** Terima cookies+localStorage dari extension, bangun `auth.json`, jalankan scraper di background |
| `/admin/scraper/emis/status` | GET | Polling status run scraper (tail log) |

**Umum**
| Endpoint | Metode | Fungsi |
|---|---|---|
| `/health` | GET | Health check (saat ini **tidak** memeriksa koneksi DB — technical debt) |

### 4.3 Model Autentikasi 3-Tingkat
| Tingkat | Mekanisme | Dipakai oleh |
|---|---|---|
| Admin | JWT (golang-jwt/v5), password bcrypt | Dashboard/admin tools |
| Internal/Service | Static bearer token per-worker (`SERVICE_TOKEN_PENDAFTARAN`, `SERVICE_TOKEN_EMIS`) | Worker-to-core-app |
| Consumer | API Key (bcrypt hash disimpan, dibandingkan via loop karena bcrypt satu arah) + scope JSON array | App presensi/perpustakaan |

Middleware: `AdminAuth` (validasi JWT via `jwtutil.ValidateToken`), `ServiceAuth` (bandingkan token statis), `APIKeyAuth` (loop bcrypt-compare ke semua `api_clients` aktif, simpan `client_scopes` ke `c.Locals`), `RequireScope`/`RequireAnyScope` (baca `client_scopes` dari Locals, 403 jika tidak cocok).

### 4.4 Hardening yang Sudah Diterapkan & Diverifikasi
- Rate limiting: 100 req/menit global + per-klien konsumen (diverifikasi: request ke-101 → HTTP 429).
- CORS: whitelist origin via `AllowOriginsFunc`, termasuk eksplisit mengizinkan prefix `moz-extension://` dan `chrome-extension://` untuk mendukung browser extension (lihat Bagian 5).
- Graceful shutdown: menangkap SIGTERM/SIGINT, drain koneksi maksimal 10 detik sebelum keluar.
- Body size limit: 10 MB (mencegah payload import yang terlalu besar membebani server).

### 4.5 Keputusan Desain Kritis (wajib dipahami sebelum mengubah kode terkait)

**4.5.1 `UpdateBiodataOnly` vs `Upsert` — JANGAN DIGABUNG KEMBALI.**
Bug historis: `SiswaService.Update()` awalnya memanggil `Upsert(..., nil, nil, nil, nil)`. Strategi upsert men-**hapus semua relasi** (wali/aktivitas_belajar/beasiswa/prestasi) lalu insert ulang dari parameter yang diberikan. Karena `Update()` biasa hanya mengirim biodata (bukan relasi), semua relasi anak tersebut **terhapus permanen** setiap kali admin mengedit biodata sederhana (misalnya memperbaiki typo nama).
**Perbaikan:** method `UpdateBiodataOnly` dibuat khusus — hanya menyentuh tabel `siswa`, tidak pernah menyentuh tabel relasi. `SiswaService.Update()` sekarang memanggil `UpdateBiodataOnly`, bukan `Upsert`.
**Aturan untuk pengembang berikutnya:** `Upsert` HANYA dipanggil dari jalur import (yang memang membawa payload relasi lengkap). Edit biodata parsial manual TIDAK PERNAH boleh memanggil `Upsert`.

**4.5.2 Field `source` di-overwrite saat re-import (trade-off yang diterima).**
Jika siswa yang sama (match by NIK) diimpor ulang dari sumber berbeda (misal awalnya dari `pendaftaran`, lalu di-scrape ulang dari EMIS), kolom `source` akan ditimpa menjadi sumber terbaru. Ini adalah trade-off yang disengaja (lihat ADR-03) — bukan bug — karena prioritas adalah data terbaru dari EMIS (sumber resmi) menang atas data pendaftaran awal.

**4.5.3 Transaksi per-record, bukan per-batch, saat import.**
Setiap baris siswa di-commit dalam transaksi sendiri. Jika baris ke-500 dari 981 gagal (misal data cacat), baris 1-499 dan 501-981 tetap tersimpan; hanya baris yang gagal dicatat di `import_log_detail` dengan status gagal. Ini disengaja agar satu baris rusak tidak membatalkan seluruh batch.

**4.5.4 Enkripsi kredensial scraper bersifat wajib, bukan opsional.**
`scraper_configs.value` untuk field sensitif (password EMIS) **harus** melalui `pkg/cryptoutil` (AES-256-GCM) sebelum disimpan. Service layer memisahkan `GetMasked()` (untuk ditampilkan ke admin, password disensor) dari `GetFull()` (hanya dipanggil oleh endpoint internal yang dikonsumsi scraper). Jangan pernah menambahkan endpoint yang mengembalikan `GetFull()` ke klien non-worker.

**4.5.5 Pencarian konsumen hanya mengembalikan field "ringkas", bukan record penuh.**
`/consumer/siswa/search` dan `/consumer/siswa/:id` sengaja membatasi field yang dikembalikan (prinsip least-privilege) — aplikasi presensi/perpustakaan tidak perlu melihat seluruh biodata keluarga, beasiswa, dsb.

---

## 5. Spesifikasi Modul: Worker Scraping EMIS (Python + Playwright)

### 5.1 Lokasi & Nama
`workers/scraping-import-emis/` (dipindahkan dari proyek referensi terpisah `data-siswa-emis`, mengikuti konvensi folder `workers/` pada proyek bank-data).

### 5.2 Alur Kerja
1. **Deteksi jumlah halaman** — galloping search + bisection terhadap halaman kesiswaan EMIS untuk menemukan total halaman data tanpa harus memuat seluruhnya (bisa di-override via `TOTAL_PAGES_OVERRIDE`/`LAST_PAGE_COUNT_OVERRIDE` di `.env` untuk menghemat request dan menghindari WAF).
2. **Per halaman**: scrape HTML → parse dengan BeautifulSoup4 → map ke struktur JSON kanonikal → **push langsung ke core-app** per halaman (bukan menunggu semua halaman selesai) via `push_page_to_core_app()`.
3. Jika push satu halaman gagal, dicatat ke `push_failures.json` (bukan menghentikan proses); dapat di-retry nanti lewat `retry_failed_push.py`.
4. `progress_backup.json` menyimpan seluruh data yang sudah di-scrape sebagai cadangan lokal independen dari core-app (untuk audit/recovery).

### 5.3 Perbedaan Mode Headless vs Interaktif
- **Mode interaktif** (dipakai sekali di awal, di laptop lokal dengan GUI): login manual ke EMIS, selesaikan captcha manual, Playwright menyimpan sesi (`storage_state`) ke `auth.json`.
- **Mode headless** (dipakai di server produksi tanpa GUI): TIDAK melakukan login — hanya memuat `auth.json` yang sudah ada dan memvalidasinya (probe halaman 1). Jika sesi EMIS yang disimpan di `auth.json` sudah kedaluwarsa dan tidak ada GUI untuk login ulang, proses akan **gagal** (lihat masalah ini diselesaikan lewat browser extension, Bagian 6) atau — sebagai technical debt yang belum diperbaiki — berpotensi **hang tanpa batas waktu** jika kode `ensure_login()` lama masih menunggu input interaktif (lihat Bagian 9, item technical debt).

### 5.4 Sentralisasi Kredensial (perubahan arsitektur penting)
**Sebelum:** `EMIS_EMAIL`/`EMIS_PASSWORD` disimpan plaintext di `.env` worker. Ini menyebabkan kredensial asli ter-paste ke chat berulang kali secara tidak sengaja (lihat insiden keamanan, Bagian 8.3).
**Sesudah:** `.env` worker hanya menyimpan `CORE_APP_URL` dan `SERVICE_TOKEN_EMIS`. Fungsi baru `fetch_config_from_core_app()` dipanggil di awal `scraper.py`, melakukan `GET /internal/scraper-config/emis-scraper` dengan header `Authorization: Bearer $SERVICE_TOKEN_EMIS`, mendapatkan `EMIS_EMAIL`, `EMIS_PASSWORD` (sudah didekripsi oleh core-app), `BASE_URL`, `ROWS_PER_PAGE`, `SCRAPER_WORKERS` — semua nilai ini TIDAK PERNAH ditulis ke disk oleh worker, hanya ada di memori proses selama runtime.

### 5.5 Variabel Konfigurasi Lokal yang Tetap di `.env` Worker
```
CORE_APP_URL=http://localhost:8080/api/v1/internal
SERVICE_TOKEN_EMIS=<token acak, dibuat via scripts/tools/generate_api_key>
IMPORT_MODE=upsert
PUSH_TO_CORE_APP=true
AUTH_FILE=auth.json
BACKUP_FILE=progress_backup.json
OUTPUT_FILE=data_siswa_emis.xlsx
TOTAL_PAGES_OVERRIDE / LAST_PAGE_COUNT_OVERRIDE   (opsional)
SCRAPER_WORKERS / ROWS_PER_PAGE                   (fallback, biasanya override dari core-app)
NAV_TIMEOUT_MS, SELECTOR_TIMEOUT_MS, DETAIL_WAIT_MS, RETRY_WAIT_MS,
POST_BACK_WAIT_MS, PROBE_TIMEOUT_MS, PROBE_RETRIES, WARMUP_TIMEOUT_MS
```

### 5.6 Entry Point Produksi
`run_scraper.sh` — mengaktifkan virtualenv, memvalidasi sesi (probe headless halaman 1), lalu menjalankan `python3 scraper.py "$MODE"`. Dipanggil **oleh core-app** sebagai background subprocess (lihat Bagian 6.5), bukan dijalankan manual via SSH.

---

## 6. Spesifikasi Modul: Browser Extension (`extension-bank-data-emis`)

### 6.1 Latar Belakang & Masalah yang Dipecahkan
Server produksi tidak memiliki GUI, sehingga tidak bisa menyelesaikan captcha EMIS atau login interaktif secara langsung di server. Dua pendekatan dipertimbangkan dan **ditolak**:
- Playwright lokal di laptop + `scp` manual hasil `auth.json` ke server — terlalu manual, perlu diulang setiap sesi EMIS kedaluwarsa.
- Xvfb/VNC di server — terlalu berat/rumit untuk deployment skala satu madrasah.

**Solusi yang dipilih:** Browser extension yang menangkap sesi EMIS yang **sudah login secara normal** di browser pengguna (tanpa perlu login ulang via Playwright sama sekali), lalu mengirim cookies + localStorage tersebut langsung ke core-app sebagai JSON. Core-app merangkai ulang struktur `storage_state` Playwright di sisi server dan langsung menjalankan scraper. (Lihat ADR-05.)

### 6.2 Manifest & Permission (Manifest V3)
- `permissions`: `cookies`, `scripting`, `activeTab`, `storage`.
- `host_permissions`: domain EMIS + `http://localhost:8080/*` + `http://127.0.0.1:8080/*`.
- `content_security_policy.extension_pages` dengan **`connect-src` eksplisit** — wajib di Firefox karena default MV3 CSP (`default-src 'none'`) memblokir semua `fetch()` dari popup kecuali domain tujuan didaftarkan secara eksplisit. (Lihat Bagian 8.4 untuk kronologi debug.)

### 6.3 Alur Fungsional (popup.js)
1. Admin mengisi username/password admin core-app sekali, disimpan di `chrome.storage.local`.
2. Klik "Jalankan" →
   a. `loginAndGetToken()` — POST `/auth/login` ke core-app, dapat JWT segar.
   b. `getCookies()` — `chrome.cookies.getAll` digabung dari seluruh domain EMIS yang relevan.
   c. `getLocalStorage(tabId)` — `chrome.scripting.executeScript` ke tab EMIS yang sedang aktif & sudah login, membaca `localStorage`.
   d. POST ke `/admin/scraper/emis/trigger-json?mode=<mode>` dengan body `{cookies, localStorage}` + header `Authorization: Bearer <jwt>`.
   e. Polling `/admin/scraper/emis/status` setiap 3 detik sampai `running: false`.

### 6.4 Known Limitations
- Harus dijalankan secara manual oleh admin yang membuka tab EMIS dan login dulu — **belum otomatis terjadwal** (lihat roadmap Fase 12).
- Belum dipaketkan sebagai `.xpi` bertanda tangan (signed) — saat ini hanya "Load Temporary Add-on" di Firefox, yang hilang setiap restart browser. Technical debt.
- Hanya diuji di Firefox desktop; Chrome/Chromium belum divalidasi end-to-end meski manifest mendukung keduanya.

### 6.5 Endpoint Server yang Mendukung Extension (baru)
- `POST /admin/scraper/emis/trigger-json` — menerima `{cookies: [...], localStorage: [...]}`, membangun `auth.json` format Playwright `storage_state` (`{cookies, origins: [{origin, localStorage}]}`) langsung di server (menggantikan script `merge_auth.py` yang sebelumnya diperlukan), lalu menjalankan `run_scraper.sh` sebagai **background subprocess** (`exec.Command` Go, output diarahkan ke file log `scraping.log`, tidak diblokir menunggu proses selesai).
- `GET /admin/scraper/emis/status` — tail log `scraping.log`, mengembalikan status `running: true/false` + isi log terbaru untuk ditampilkan di popup extension.

---

## 7. Keamanan

### 7.1 Model Autentikasi
Lihat Bagian 4.3 (3-tingkat: Admin JWT, Service Token, Consumer API Key berskop).

### 7.2 Enkripsi Data Sensitif
AES-256-GCM (`pkg/cryptoutil`) untuk kredensial EMIS di tabel `scraper_configs`. Kunci enkripsi (`ENCRYPTION_KEY`) disimpan di `.env` core-app (bukan di database), dipisahkan dari ciphertext.

### 7.3 Insiden Kebocoran Kredensial — Catatan Wajib Baca
Selama pengembangan, kredensial asli (password EMIS `Pupung123.@`, sesi/JWT EMIS hidup, dan password admin yang sama dengan password root MySQL) **ter-paste sebagai plaintext ke dalam chat setidaknya 4 kali** (dump `auth.json`, dua kali paste `.env` berisi `EMIS_PASSWORD`, dan satu kali body request `POST scraper-config` berisi password asli). Setiap kali terjadi, peringatan eksplisit diberikan untuk segera merotasi kredensial tersebut.

**Aturan keras untuk AI/engineer berikutnya:**
- **Jangan pernah** menampilkan ulang (echo/cat/log) isi file kredensial secara verbatim di respons atau log yang bisa dibaca orang lain, bahkan untuk tujuan debugging.
- Jika pengguna menempelkan kredensial asli ke percakapan, **segera** ingatkan untuk merotasi kredensial tersebut sesegera mungkin, dan hindari mengulang nilai tersebut dalam respons berikutnya.
- Simpan kredensial hanya di: (a) `.env` yang di-gitignore, untuk kredensial infrastruktur (DB, JWT secret, encryption key, service token); atau (b) kolom terenkripsi di database untuk kredensial pihak ketiga yang perlu diakses otomatis oleh worker (seperti EMIS).

### 7.4 Status Validasi Auth Middleware
`AdminAuth` dan `APIKeyAuth` awalnya adalah placeholder tanpa validasi nyata — ditemukan dan diperbaiki setelah instruksi eksplisit "jangan pernah meninggalkan bug yang krusial". `AdminAuth` sekarang memvalidasi JWT penuh; `APIKeyAuth` melakukan bcrypt-compare loop terhadap seluruh `api_clients` aktif (karena bcrypt tidak bisa di-query langsung by hash).

---

## 8. Infrastruktur & Deployment

### 8.1 Keputusan No-Docker (mengikat)
Deployment menggunakan **PM2**, bukan Docker/container apa pun. Ini permintaan eksplisit pemilik proyek. Jangan memperkenalkan Docker/docker-compose di tahap mana pun tanpa persetujuan ulang.

### 8.2 Konfigurasi PM2 (`ecosystem.config.js`, root proyek)
- App `core-app`: mode `fork`, `cwd: "./backend"`, script binary hasil `go build`, `autorestart: true`, `max_memory_restart: "300M"`.
- **Catatan path penting:** `out_file`/`error_file` harus relatif terhadap `cwd`, BUKAN digabung manual dengan `./backend/...` lagi — kesalahan ini pernah membuat PM2 membuat folder `backend/backend/logs` secara tidak sengaja.
- Startup persistence: `pm2 startup` (systemd) + `pm2 save` sudah dikonfigurasi dan diverifikasi bertahan setelah reboot.
- Build/deploy ulang: `scripts/rebuild.sh` (go build → pm2 restart → tampilkan status). **Setelah mengubah `.env`, wajib jalankan `pm2 restart core-app --update-env`** — restart biasa tidak membaca ulang environment variables.

### 8.3 Variabel Environment (`backend/.env`)
```
DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME
JWT_SECRET
ENCRYPTION_KEY
SERVICE_TOKEN_PENDAFTARAN
SERVICE_TOKEN_EMIS
CORS_ALLOWED_ORIGINS
SCRAPER_DIR          (path absolut ke folder worker; WAJIB absolut, bukan relatif — lihat Bagian 8.4)
```

### 8.4 Bug Path Historis — `SCRAPER_DIR` harus absolut
Handler trigger awalnya memakai path relatif (`"../workers/scraping-import-emis"`) yang bergantung pada *current working directory* saat proses dijalankan. `go run` dari root proyek berbeda cwd dibanding PM2 (`cwd: "./backend"`), menyebabkan error `fork/exec ...: no such file or directory` di produksi meski berjalan normal saat development lokal. **Perbaikan permanen:** `SCRAPER_DIR` di `.env` sebagai path absolut, dibaca lewat `cfg.ScraperDir()`. Jangan kembalikan ke path relatif.

---

## 9. Status Implementasi (per 5 Oktober 2026)

### 9.1 Selesai & Teruji (Done & Tested)
1. CRUD penuh siswa + relasi (wali/aktivitas/beasiswa/prestasi) — admin API.
2. Import JSON (insert & upsert) dengan transaksi per-record dan `import_log` lengkap.
3. Export JSON & Excel (3 sheet) data siswa.
4. Auth 3-tingkat (Admin JWT, Service Token, Consumer API Key berskop) — divalidasi positif & negatif.
5. Endpoint `activity_log` ringan untuk presensi/perpustakaan.
6. Admin CRUD API client (create/list/update/regenerate key/delete) — diuji lengkap.
7. Hardening: rate limit (100/menit, terverifikasi 429 di request ke-101), CORS whitelist + extension origin, graceful shutdown (SIGTERM, drain 10s), body limit 10MB.
8. PM2 process management + boot persistence (systemd) — terverifikasi bertahan restart.
9. Sentralisasi & enkripsi AES-GCM kredensial scraper EMIS di database, diambil via endpoint internal terautentikasi.
10. Browser extension Firefox end-to-end: capture sesi EMIS → auto-login admin → trigger scraper background → polling log.
11. **Hasil scraping penuh terverifikasi: 981/981 siswa berhasil di-scrape dan di-push ke core-app, 0 kegagalan (`is_failed=0` untuk semua, `push_failures.json` kosong).**

### 9.2 Technical Debt (belum dikerjakan, tercatat untuk tindak lanjut)
| # | Item | Risiko jika diabaikan |
|---|---|---|
| 1 | Import queue asinkron (saat ini sinkron per-request) | Request import besar bisa timeout di sisi klien |
| 2 | Admin audit log (siapa mengubah apa) | Tidak ada jejak akuntabilitas untuk perubahan data manual |
| 3 | Rate limit dinamis per-klien (saat ini flat 100/menit) | Klien dengan kebutuhan beban tinggi bisa terblokir tidak adil |
| 4 | `ensure_login()` scraper bisa hang tanpa batas waktu jika dijalankan non-interaktif tanpa `auth.json` valid | Proses worker bisa menggantung selamanya di cron/PM2 tanpa timeout |
| 5 | `/health` tidak memeriksa koneksi database | Core-app bisa terlihat "sehat" walau DB down |
| 6 | Tidak ada dokumentasi OpenAPI/Swagger | Integrasi app presensi/perpustakaan bergantung pada dokumentasi manual/tacit knowledge |
| 7 | Tidak ada automated test suite (semua pengujian sejauh ini manual via curl) | Regresi tidak terdeteksi otomatis |
| 8 | Extension belum dipaketkan sebagai `.xpi` bertanda tangan | Extension hilang setiap restart browser, tidak bisa didistribusikan resmi |
| 9 | Trigger scraping masih manual (klik admin), belum terjadwal otomatis | Data EMIS bisa menjadi usang jika admin lupa menjalankan |
| 10 | Validasi input NIK/NISN belum ketat (format, checksum) | Data cacat bisa masuk dari sumber manual/import |

---

## 10. Rencana Tahap Berikutnya (Roadmap)

### Fase 9 — `scraping-export-emis` (Sinkronisasi balik ke EMIS)
**Status: desain belum final — ini adalah pekerjaan rintisan untuk AI/engineer berikutnya.**
Mekanisme penulisan balik ke EMIS (push data dari core-app kembali ke portal) **belum ada kepastian** apakah EMIS bahkan menerima input otomatis (form submission terstruktur) atau hanya bisa diisi manual. Langkah pertama yang disarankan: audit ulang portal EMIS untuk menentukan apakah ada form yang bisa diisi otomatis via Playwright (mirip pendekatan scraping-import), dan apakah ada risiko kebijakan/ToS yang harus dikonfirmasi ke pemilik proyek sebelum membangun otomasi tulis ke sistem pemerintah.

### Fase 10 — Perluasan Endpoint Tulis Presensi/Perpustakaan
**Wajib konfirmasi ulang dengan pemilik proyek sebelum mengerjakan.** Keputusan arsitektur saat ini (Bagian 2.3 & ADR-02) **secara sengaja** membatasi core-app hanya mencatat `activity_log` ringan, bukan data bisnis penuh presensi/peminjaman buku. Jangan memperluas skema untuk menyimpan data detail presensi/perpustakaan tanpa konfirmasi eksplisit bahwa keputusan ini berubah — ini bukan gap yang harus "disempurnakan" secara default.

### Fase 11 — Dashboard Nuxt
Belum dimulai. Akan menjadi antarmuka web admin (menggantikan kebutuhan curl manual untuk operasi CRUD/import/export/lihat log).

### Fase 12 — Hardening & Skalabilitas Lanjutan
Mencakup seluruh item technical debt Bagian 9.2, plus: penjadwalan otomatis trigger scraping (menggantikan klik manual admin), dan evaluasi apakah volume data/traffic di masa depan memerlukan penyesuaian arsitektur (connection pooling, caching, dsb).

---

## 11. Kebutuhan Non-Fungsional

- **Keandalan:** Proses import/scraping harus tahan terhadap kegagalan parsial (sudah dipenuhi via transaksi per-record & `push_failures.json`).
- **Keamanan:** Tidak ada kredensial pihak ketiga dalam bentuk plaintext yang persisten di disk mana pun (sudah dipenuhi via enkripsi AES-GCM + sentralisasi).
- **Portabilitas:** Tidak bergantung pada container runtime (Docker) — hanya bergantung pada Go toolchain, Node/PM2, Python/virtualenv, dan MySQL yang terinstal langsung di host.
- **Kompatibilitas:** Core-app harus tetap bisa diakses dari origin `moz-extension://`/`chrome-extension://` selama extension masih menjadi mekanisme trigger utama scraping.
- **Auditabilitas:** Semua import harus tercatat di `import_log`/`import_log_detail`; semua akses klien konsumen tercatat di `api_client_access_log`.

---

## 12. Glosarium

| Istilah | Arti |
|---|---|
| SSOT | Single Source of Truth — satu sumber kebenaran data, di sini adalah core-app |
| EMIS | Education Management Information System — portal data kesiswaan Kementerian Agama |
| Upsert | Insert jika belum ada, update jika sudah ada (match by NIK) |
| Service Token | Token statis untuk autentikasi worker-ke-core-app (bukan JWT) |
| Scope | Hak akses granular yang dilampirkan pada API Key konsumen (misal `siswa:read`) |
| `storage_state` | Format Playwright untuk menyimpan sesi browser (cookies + localStorage) agar bisa dipakai ulang tanpa login interaktif |
| Activity log | Jejak ringan aktivitas dari aplikasi konsumen, bukan data bisnis penuh |

---

## 13. Appendix — Architecture Decision Records (ADR)

**ADR-01: Tidak menggunakan Docker, memakai PM2.**
Keputusan eksplisit pemilik proyek. Alasan: kesederhanaan operasional untuk skala satu madrasah, tim tidak familiar/tidak ingin mengelola container di server produksi yang sudah ada.

**ADR-02: Core-app hanya menyimpan `activity_log` ringan dari presensi/perpustakaan, bukan data bisnis penuh.**
Alasan: menghindari duplikasi sumber kebenaran — data presensi/peminjaman buku tetap dimiliki oleh aplikasi masing-masing; core-app hanya perlu tahu "siswa X pernah berinteraksi dengan sistem Y pada waktu Z" untuk keperluan audit lintas sistem, bukan menjadi gudang data kedua.

**ADR-03: Field `source` pada tabel `siswa` di-overwrite oleh sumber import terbaru (bukan diakumulasi/digabung).**
Trade-off yang diterima: EMIS dianggap sumber paling otoritatif begitu tersedia, sehingga re-import dari EMIS menimpa `source` sebelumnya (misal dari `pendaftaran`). Konsekuensi: histori asal-usul data per-field tidak dipertahankan — hanya sumber terakhir yang tercatat.

**ADR-04: Kredensial EMIS disentralisasi & dienkripsi di database core-app, bukan di `.env` worker.**
Dipicu langsung oleh insiden kebocoran kredensial berulang (Bagian 7.3). Efek samping positif: rotasi kredensial sekarang terjadi di satu tempat (lewat admin endpoint), tidak perlu redeploy worker.

**ADR-05: Browser extension dipilih atas Xvfb/VNC atau scp manual untuk mengatasi "server tanpa GUI + captcha EMIS".**
Alasan: extension menangkap sesi yang *sudah* terautentikasi secara alami oleh pengguna di browser mereka sendiri (tidak perlu captcha sama sekali dalam alur otomatis), jauh lebih ringan daripada menjalankan virtual display server, dan tidak memerlukan akses SSH langsung ke server produksi untuk memicu scraping.

---

## 14. Catatan Khusus untuk AI/Engineer Selanjutnya

1. **Jangan** menggabungkan kembali `UpdateBiodataOnly` ke dalam `Upsert` — ini akan menghidupkan kembali bug penghapusan data relasi (Bagian 4.5.1).
2. **Jangan** menampilkan ulang isi file kredensial (`.env`, `auth.json`, dsb.) secara verbatim di output/log untuk tujuan debugging apa pun (Bagian 7.3).
3. **Jangan** memperkenalkan Docker di tahap mana pun tanpa persetujuan eksplisit ulang dari pemilik proyek (ADR-01).
4. **Jangan** memperluas skema data untuk menyimpan detail presensi/perpustakaan penuh tanpa konfirmasi ulang — ini adalah keputusan arsitektur sengaja, bukan kekurangan (ADR-02).
5. Endpoint baru yang berinteraksi dengan data sensitif (kredensial, biodata lengkap) **harus** mengikuti pola middleware yang sudah ada (`AdminAuth`/`ServiceAuth`/`APIKeyAuth` + `RequireScope`), jangan membuat jalur auth baru.
6. Migrasi database baru harus reversible dan diuji `up → down → up` sebelum dianggap selesai, mengikuti pola 10 migrasi yang sudah ada.
7. Path ke resource eksternal (seperti folder worker) harus selalu dikonfigurasi sebagai **path absolut** via `.env`, tidak relatif terhadap cwd proses (lihat Bagian 8.4).
8. Sebelum memulai Fase 9 (`scraping-export-emis`), pastikan dulu secara faktual apakah EMIS mengizinkan/mendukung input otomatis — jangan asumsikan ini layak tanpa verifikasi.

**— Akhir Dokumen —**
