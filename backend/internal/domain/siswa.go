package domain

import "time"

type Siswa struct {
	ID                    int64      `json:"id"`
	UUID                  string     `json:"uuid"`
	NIK                   *string    `json:"nik"`
	NISN                  *string    `json:"nisn"`
	KIP                   *string    `json:"kip"`
	Nama                  string     `json:"nama"`
	TempatLahir           *string    `json:"tempat_lahir"`
	TanggalLahir          *time.Time `json:"tanggal_lahir"`
	JenisKelamin          *string    `json:"jenis_kelamin"`
	Agama                 *string    `json:"agama"`
	JumlahSaudara         *int       `json:"jumlah_saudara"`
	AnakKe                *int       `json:"anak_ke"`
	Hobi                  *string    `json:"hobi"`
	CitaCita              *string    `json:"cita_cita"`
	NoHandphone           *string    `json:"no_handphone"`
	Email                 *string    `json:"email"`
	YangMembiayai         *string    `json:"yang_membiayai"`
	KebutuhanDisabilitas  *string    `json:"kebutuhan_disabilitas"`
	KebutuhanKhusus       *string    `json:"kebutuhan_khusus"`
	Alamat                *string    `json:"alamat"`
	StatusTempatTinggal   *string    `json:"status_tempat_tinggal"`
	JarakRumahMadrasah    *string    `json:"jarak_rumah_madrasah"`
	WaktuTempuh           *string    `json:"waktu_tempuh"`
	Transportasi          *string    `json:"transportasi"`
	Fingerprint           *string    `json:"fingerprint"`
	RFID                  *string    `json:"rfid"`
	Source                string     `json:"source"`
	SourcePage            *int       `json:"source_page"`
	SourceRow             *int       `json:"source_row"`
	IsFailed              bool       `json:"is_failed"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
	RawJSON               *string    `json:"raw_json,omitempty"`
}

// Payload yang diterima dari import JSON (sesuai struktur asli kamu)
type SiswaImportPayload struct {
	Siswa            map[string]string   `json:"siswa"`
	Ayah             map[string]string   `json:"ayah"`
	Ibu              map[string]string   `json:"ibu"`
	Wali             map[string]string   `json:"wali"`
	AktivitasBelajar []map[string]string `json:"aktivitas_belajar"`
	Beasiswa         []map[string]string `json:"beasiswa"`
	Prestasi         []map[string]string `json:"prestasi"`
	Page             int                 `json:"_page"`
	Row              int                 `json:"_row"`
	Failed           bool                `json:"_failed"`
}
