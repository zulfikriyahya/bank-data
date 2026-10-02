package service

import (
	"encoding/json"

	"bank-data/backend/internal/domain"
	"bank-data/backend/internal/repository"
	"bank-data/backend/pkg/excelgen"

	"github.com/xuri/excelize/v2"
)

type ExportService struct {
	siswaRepo repository.SiswaRepository
}

func NewExportService(siswaRepo repository.SiswaRepository) *ExportService {
	return &ExportService{siswaRepo: siswaRepo}
}

func (s *ExportService) ExportToJSON(filter repository.ExportFilter) ([]byte, error) {
	data, err := s.siswaRepo.FindAllForExport(filter)
	if err != nil {
		return nil, err
	}

	// bentuk ulang supaya outputnya mirip struktur import asli (siswa, ayah, ibu, wali, dst)
	// supaya file hasil export BISA dipakai lagi sebagai file import (round-trip)
	type exportRow struct {
		Siswa            domain.Siswa                `json:"siswa"`
		Ayah             *domain.WaliSiswa            `json:"ayah,omitempty"`
		Ibu              *domain.WaliSiswa            `json:"ibu,omitempty"`
		Wali             *domain.WaliSiswa            `json:"wali,omitempty"`
		AktivitasBelajar []domain.AktivitasBelajar    `json:"aktivitas_belajar"`
		Beasiswa         []domain.Beasiswa            `json:"beasiswa"`
		Prestasi         []domain.Prestasi            `json:"prestasi"`
	}

	rows := make([]exportRow, 0, len(data))
	for _, d := range data {
		rows = append(rows, exportRow{
			Siswa:            d.Siswa,
			Ayah:             d.Ayah,
			Ibu:              d.Ibu,
			Wali:             d.Wali,
			AktivitasBelajar: d.AktivitasBelajar,
			Beasiswa:         d.Beasiswa,
			Prestasi:         d.Prestasi,
		})
	}

	return json.MarshalIndent(rows, "", "  ")
}

func (s *ExportService) ExportToExcel(filter repository.ExportFilter) (*excelize.File, error) {
	data, err := s.siswaRepo.FindAllForExport(filter)
	if err != nil {
		return nil, err
	}

	return excelgen.GenerateSiswaExcel(data)
}
