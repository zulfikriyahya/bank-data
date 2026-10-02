package domain

import "time"

type WaliSiswa struct {
	ID                  int64      `json:"id"`
	SiswaID             int64      `json:"siswa_id"`
	Role                string     `json:"role"` // ayah, ibu, wali
	NIK                 *string    `json:"nik"`
	NamaLengkap         *string    `json:"nama_lengkap"`
	TempatLahir         *string    `json:"tempat_lahir"`
	TanggalLahir        *time.Time `json:"tanggal_lahir"`
	Status              *string    `json:"status"`
	PendidikanTerakhir  *string    `json:"pendidikan_terakhir"`
	PekerjaanUtama      *string    `json:"pekerjaan_utama"`
	Domisili            *string    `json:"domisili"`
	NoHandphone         *string    `json:"no_handphone"`
	Penghasilan         *string    `json:"penghasilan"`
	Alamat              *string    `json:"alamat"`
	StatusTempatTinggal *string    `json:"status_tempat_tinggal"`
}
