package domain

type SiswaExportData struct {
	Siswa            Siswa
	Ayah             *WaliSiswa
	Ibu              *WaliSiswa
	Wali             *WaliSiswa
	AktivitasBelajar []AktivitasBelajar
	Beasiswa         []Beasiswa
	Prestasi         []Prestasi
}
