package domain

type Prestasi struct {
	ID           int64  `json:"id"`
	SiswaID      int64  `json:"siswa_id"`
	NamaPrestasi string `json:"nama_prestasi"`
	Tingkat      string `json:"tingkat"`
	Tahun        string `json:"tahun"`
	Keterangan   string `json:"keterangan"`
}
