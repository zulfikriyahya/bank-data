package domain

type Beasiswa struct {
	ID           int64  `json:"id"`
	SiswaID      int64  `json:"siswa_id"`
	NamaBeasiswa string `json:"nama_beasiswa"`
	Tahun        string `json:"tahun"`
	Keterangan   string `json:"keterangan"`
}
