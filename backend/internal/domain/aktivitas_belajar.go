package domain

import "time"

type AktivitasBelajar struct {
	ID                  int64      `json:"id"`
	SiswaID             int64      `json:"siswa_id"`
	TahunAjaranSemester *string    `json:"tahun_ajaran_semester"`
	TanggalMulaiMasuk   *time.Time `json:"tanggal_mulai_masuk"`
	NSMNamaLembaga      *string    `json:"nsm_nama_lembaga"`
	TingkatKelompok     *string    `json:"tingkat_kelompok"`
	Jurusan             *string    `json:"jurusan"`
	StatusKeaktifan     *string    `json:"status_keaktifan"`
	Keterangan          *string    `json:"keterangan"`
}
