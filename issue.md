# Master Data Management (MDM)
## requirements
### BACKEND
- golang[fiber]
- mysql
- ...

### FRONTEND
- NUXTJS 3 /VUE
- TAILWINDCSS
- tanstack
- pinia (state)

## feature
- import data json [insert data, upsert data]
- export data [json, excel]
- ...

## struktur data:
```json
[
  {
    "siswa": {
      "NIK": "3601216104140003",
      "NISN": "3149022989",
      "KIP": "-",
      "TEMPAT LAHIR": "Pandeglang",
      "TANGGAL LAHIR": "21 April 2014",
      "JENIS KELAMIN": "Perempuan",
      "AGAMA": "Islam",
      "JUMLAH SAUDARA": "2",
      "ANAK KE": "2",
      "HOBI": "Menulis",
      "CITA-CITA": "Guru/Dosen",
      "NO. HANDPHONE": "0",
      "ALAMAT EMAIL SISWA": "-",
      "YANG MEMBIAYAI SEKOLAH": "Orang Tua",
      "KEBUTUHAN DISABILITAS": "Tidak Ada",
      "KEBUTUHAN KHUSUS": "Tidak Ada",
      "ALAMAT": "Komp. Graha Cipacung, Kel. SARUNI, Kec. MAJASARI, KABUPATEN PANDEGLANG, BANTEN, 42216",
      "STATUS TEMPAT TINGGAL": "Tinggal dengan Orangtua/Wali",
      "JARAK TEMPAT TINGGAL - MADRASAH": "Kurang dari 5 km",
      "WAKTU TEMPUH": "10-19 menit",
      "TRANSPORTASI KE SEKOLAH": "Sepeda Motor",
      "NAMA": "ADZKIYA UFAIRANISSA FATHURROHMAN",
      "FINGERPRINT":"-",
      "RFID":"0231562854",
      "AVATAR":""
    },
    "ayah": {
      "NIK": "3601210912820001",
      "NAMA LENGKAP": "NANANG FATHURROHMAN",
      "TEMPAT LAHIR": "Serang",
      "TANGGAL LAHIR": "09 Desember 1982",
      "STATUS": "Masih Hidup",
      "PENDIDIKAN TERAKHIR": "D4/S1",
      "PEKERJAAN UTAMA": "Guru/Dosen",
      "DOMISILI": "Dalam Negeri",
      "NO HANDPHONE": "6285226912333",
      "PENGHASILAN RATA-RATA PER BULAN (Rp)": "2.500.001 - 3.500.000",
      "ALAMAT": "Komp. Graha Cipacung, Kel. SARUNI, Kec. MAJASARI, KABUPATEN PANDEGLANG, BANTEN",
      "STATUS TEMPAT TINGGAL": "Milik Sendiri"
    },
    "ibu": {
      "NIK": "3601216605820001",
      "NAMA LENGKAP": "RATU IKHDA HAYATI",
      "TEMPAT LAHIR": "Pandeglang",
      "TANGGAL LAHIR": "26 Mei 1982",
      "STATUS": "Masih Hidup",
      "PENDIDIKAN TERAKHIR": "D4/S1",
      "PEKERJAAN UTAMA": "Tidak Bekerja",
      "DOMISILI": "Dalam Negeri",
      "NO HANDPHONE": "-",
      "PENGHASILAN RATA-RATA PER BULAN (Rp)": "2.500.001 - 3.500.000",
      "ALAMAT": "Komp. Graha Cipacung, Kel. SARUNI, Kec. MAJASARI, KABUPATEN PANDEGLANG, BANTEN",
      "STATUS TEMPAT TINGGAL": "Milik Sendiri"
    },
    "wali": {
      "NIK": "3601210912820001",
      "NAMA LENGKAP": "NANANG FATHURROHMAN",
      "TEMPAT LAHIR": "Serang",
      "TANGGAL LAHIR": "09 Desember 1982",
      "STATUS": "Sama dengan ayah kandung",
      "PENDIDIKAN TERAKHIR": "D4/S1",
      "PEKERJAAN UTAMA": "Guru/Dosen",
      "DOMISILI": "Dalam Negeri",
      "NO HANDPHONE": "6285226912333",
      "PENGHASILAN RATA-RATA PER BULAN (Rp)": "2.500.001 - 3.500.000",
      "ALAMAT": "Komp. Graha Cipacung, Kel. SARUNI, Kec. MAJASARI, KABUPATEN PANDEGLANG, BANTEN",
      "STATUS TEMPAT TINGGAL": "Milik Sendiri"
    },
    "aktivitas_belajar": [
      {
        "TAHUN AJARAN - SEMESTER": "",
        "TANGGAL MULAI MASUK": "",
        "NSM - NAMA LEMBAGA": "",
        "TINGKAT / KELOMPOK": "",
        "JURUSAN": "",
        "STATUS KEAKTIFAN": "",
        "KETERANGAN": ""
      },
      {
        "TAHUN AJARAN - SEMESTER": "",
        "TANGGAL MULAI MASUK": "",
        "NSM - NAMA LEMBAGA": "",
        "TINGKAT / KELOMPOK": "",
        "JURUSAN": "",
        "STATUS KEAKTIFAN": "",
        "KETERANGAN": ""
      },
      {
        "TAHUN AJARAN - SEMESTER": "",
        "TANGGAL MULAI MASUK": "",
        "NSM - NAMA LEMBAGA": "",
        "TINGKAT / KELOMPOK": "",
        "JURUSAN": "",
        "STATUS KEAKTIFAN": "",
        "KETERANGAN": ""
      }
    ],
    "beasiswa": [],
    "prestasi": [],
    "_page": 1,
    "_row": 1,
    "_failed": false
  }
]
```
