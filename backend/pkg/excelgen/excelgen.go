package excelgen

import (
	"bank-data/backend/internal/domain"
	"fmt"
	"time"

	"github.com/xuri/excelize/v2"
)

// GenerateSiswaExcel - bikin file Excel dengan 2 sheet: Siswa (ringkas) dan Orang Tua/Wali (detail)
func GenerateSiswaExcel(data []domain.SiswaExportData) (*excelize.File, error) {
	f := excelize.NewFile()

	if err := writeSheetSiswa(f, data); err != nil {
		return nil, err
	}
	if err := writeSheetOrtuWali(f, data); err != nil {
		return nil, err
	}
	if err := writeSheetAktivitasBelajar(f, data); err != nil {
		return nil, err
	}

	f.DeleteSheet("Sheet1") // hapus default sheet kosong
	f.SetActiveSheet(0)

	return f, nil
}

func writeSheetSiswa(f *excelize.File, data []domain.SiswaExportData) error {
	sheet := "Siswa"
	f.NewSheet(sheet)

	headers := []string{
		"NIK", "NISN", "KIP", "Nama", "Tempat Lahir", "Tanggal Lahir", "Jenis Kelamin",
		"Agama", "Jumlah Saudara", "Anak Ke", "Hobi", "Cita-Cita", "No HP", "Email",
		"Yang Membiayai", "Disabilitas", "Kebutuhan Khusus", "Alamat", "Status Tempat Tinggal",
		"Jarak ke Madrasah", "Waktu Tempuh", "Transportasi", "RFID", "Sumber Data",
	}
	for i, h := range headers {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetCellValue(sheet, col+"1", h)
	}

	for rowIdx, d := range data {
		row := rowIdx + 2
		s := d.Siswa

		values := []interface{}{
			derefStr(s.NIK), derefStr(s.NISN), derefStr(s.KIP), s.Nama,
			derefStr(s.TempatLahir), formatDate(s.TanggalLahir), derefStr(s.JenisKelamin),
			derefStr(s.Agama), derefInt(s.JumlahSaudara), derefInt(s.AnakKe),
			derefStr(s.Hobi), derefStr(s.CitaCita), derefStr(s.NoHandphone), derefStr(s.Email),
			derefStr(s.YangMembiayai), derefStr(s.KebutuhanDisabilitas), derefStr(s.KebutuhanKhusus),
			derefStr(s.Alamat), derefStr(s.StatusTempatTinggal),
			derefStr(s.JarakRumahMadrasah), derefStr(s.WaktuTempuh), derefStr(s.Transportasi),
			derefStr(s.RFID), s.Source,
		}

		for i, v := range values {
			col, _ := excelize.ColumnNumberToName(i + 1)
			f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, row), v)
		}
	}

	return nil
}

func writeSheetOrtuWali(f *excelize.File, data []domain.SiswaExportData) error {
	sheet := "Orang Tua & Wali"
	f.NewSheet(sheet)

	headers := []string{
		"NIK Siswa", "Nama Siswa", "Role", "NIK", "Nama Lengkap", "Tempat Lahir", "Tanggal Lahir",
		"Status", "Pendidikan Terakhir", "Pekerjaan", "No HP", "Penghasilan", "Alamat",
	}
	for i, h := range headers {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetCellValue(sheet, col+"1", h)
	}

	row := 2
	for _, d := range data {
		for _, w := range []*domain.WaliSiswa{d.Ayah, d.Ibu, d.Wali} {
			if w == nil {
				continue
			}

			values := []interface{}{
				derefStr(d.Siswa.NIK), d.Siswa.Nama, w.Role,
				derefStr(w.NIK), derefStr(w.NamaLengkap), derefStr(w.TempatLahir), formatDate(w.TanggalLahir),
				derefStr(w.Status), derefStr(w.PendidikanTerakhir), derefStr(w.PekerjaanUtama),
				derefStr(w.NoHandphone), derefStr(w.Penghasilan), derefStr(w.Alamat),
			}

			for i, v := range values {
				col, _ := excelize.ColumnNumberToName(i + 1)
				f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, row), v)
			}
			row++
		}
	}

	return nil
}

func writeSheetAktivitasBelajar(f *excelize.File, data []domain.SiswaExportData) error {
	sheet := "Riwayat Belajar"
	f.NewSheet(sheet)

	headers := []string{
		"NIK Siswa", "Nama Siswa", "Tahun Ajaran", "Tanggal Masuk", "Lembaga",
		"Tingkat/Kelompok", "Jurusan", "Status Keaktifan", "Keterangan",
	}
	for i, h := range headers {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetCellValue(sheet, col+"1", h)
	}

	row := 2
	for _, d := range data {
		for _, a := range d.AktivitasBelajar {
			values := []interface{}{
				derefStr(d.Siswa.NIK), d.Siswa.Nama,
				derefStr(a.TahunAjaranSemester), formatDate(a.TanggalMulaiMasuk), derefStr(a.NSMNamaLembaga),
				derefStr(a.TingkatKelompok), derefStr(a.Jurusan), derefStr(a.StatusKeaktifan), derefStr(a.Keterangan),
			}

			for i, v := range values {
				col, _ := excelize.ColumnNumberToName(i + 1)
				f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, row), v)
			}
			row++
		}
	}

	return nil
}

// ===== helper =====

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefInt(i *int) interface{} {
	if i == nil {
		return ""
	}
	return *i
}

func formatDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}
