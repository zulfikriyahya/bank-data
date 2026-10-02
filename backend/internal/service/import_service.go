package service

import (
	"fmt"
	"log"

	"bank-data/backend/internal/domain"
	"bank-data/backend/internal/repository"
	"bank-data/backend/pkg/jsonparser"
)

type ImportService struct {
	siswaRepo     repository.SiswaRepository
	importLogRepo repository.ImportLogRepository
}

func NewImportService(siswaRepo repository.SiswaRepository, importLogRepo repository.ImportLogRepository) *ImportService {
	return &ImportService{
		siswaRepo:     siswaRepo,
		importLogRepo: importLogRepo,
	}
}

// ProcessImport — proses sync untuk sekumpulan payload (versi awal, belum async queue)
func (s *ImportService) ProcessImport(jobID, source, mode string, payloads []domain.SiswaImportPayload) (*domain.ImportResult, error) {
	// 1. catat job dimulai
	importLog := &domain.ImportLog{
		JobID:       jobID,
		Source:      source,
		Mode:        mode,
		Status:      "processing",
		TotalRows:   len(payloads),
		TriggeredBy: "system",
	}

	logID, err := s.importLogRepo.CreateJob(importLog)
	if err != nil {
		return nil, fmt.Errorf("gagal mencatat job: %w", err)
	}

	result := &domain.ImportResult{
		JobID:     jobID,
		TotalRows: len(payloads),
		Details:   make([]domain.ImportResultRow, 0, len(payloads)),
	}

	// 2. proses tiap payload satu per satu
	// catatan: tiap payload punya transaction sendiri (lihat repo),
	// jadi kalau 1 baris gagal, baris lain tetap lanjut diproses
	for _, payload := range payloads {
		row := s.processSingleRow(logID, source, mode, payload)
		result.Details = append(result.Details, row)

		switch row.Status {
		case "success":
			result.SuccessRows++
		case "failed":
			result.FailedRows++
		case "duplicate":
			result.DuplicateRows++
		}
	}

	// 3. update status job final
	finalStatus := "completed"
	if result.FailedRows == result.TotalRows {
		finalStatus = "failed"
	}

	if err := s.importLogRepo.UpdateJobStatus(jobID, finalStatus, result.SuccessRows, result.FailedRows, result.DuplicateRows); err != nil {
		log.Printf("warning: gagal update status job %s: %v", jobID, err)
	}

	return result, nil
}

// processSingleRow — mapping + simpan 1 record siswa, tidak pernah panic/stop proses lain
func (s *ImportService) processSingleRow(logID int64, source, mode string, payload domain.SiswaImportPayload) domain.ImportResultRow {
	resultRow := domain.ImportResultRow{
		Row:  payload.Row,
		Page: payload.Page,
	}

	// validasi minimal: nama wajib ada
	namaSiswa := payload.Siswa["NAMA"]
	if namaSiswa == "" {
		resultRow.Status = "failed"
		resultRow.Message = "field NAMA siswa kosong, baris dilewati"
		s.logDetail(logID, payload.Page, payload.Row, nil, "failed", resultRow.Message)
		return resultRow
	}

	// mapping payload mentah -> domain struct
	siswa := jsonparser.MapToSiswa(payload.Siswa, source, payload.Page, payload.Row, payload.Failed, payload)

	var wali []domain.WaliSiswa
	if len(payload.Ayah) > 0 && hasContent(payload.Ayah) {
		wali = append(wali, jsonparser.MapToWali(payload.Ayah, "ayah"))
	}
	if len(payload.Ibu) > 0 && hasContent(payload.Ibu) {
		wali = append(wali, jsonparser.MapToWali(payload.Ibu, "ibu"))
	}
	if len(payload.Wali) > 0 && hasContent(payload.Wali) {
		wali = append(wali, jsonparser.MapToWali(payload.Wali, "wali"))
	}

	aktivitas := jsonparser.MapToAktivitasBelajar(payload.AktivitasBelajar)
	beasiswa := jsonparser.MapToBeasiswa(payload.Beasiswa)
	prestasi := jsonparser.MapToPrestasi(payload.Prestasi)

	// simpan sesuai mode
	if mode == "insert" {
		// cek duplikat dulu kalau mode insert murni (tidak boleh dobel NIK)
		if siswa.NIK != nil {
			existing, err := s.siswaRepo.FindByNIK(*siswa.NIK)
			if err == nil && existing != nil {
				resultRow.Status = "duplicate"
				resultRow.Message = fmt.Sprintf("NIK %s sudah terdaftar (id: %d)", *siswa.NIK, existing.ID)
				s.logDetail(logID, payload.Page, payload.Row, &existing.ID, "duplicate", resultRow.Message)
				return resultRow
			}
		}

		siswaID, err := s.siswaRepo.Insert(siswa, wali, aktivitas, beasiswa, prestasi)
		if err != nil {
			resultRow.Status = "failed"
			resultRow.Message = err.Error()
			s.logDetail(logID, payload.Page, payload.Row, nil, "failed", err.Error())
			return resultRow
		}

		resultRow.Status = "success"
		s.logDetail(logID, payload.Page, payload.Row, &siswaID, "success", "")
		return resultRow
	}

	// mode upsert
	siswaID, isNew, err := s.siswaRepo.Upsert(siswa, wali, aktivitas, beasiswa, prestasi)
	if err != nil {
		resultRow.Status = "failed"
		resultRow.Message = err.Error()
		s.logDetail(logID, payload.Page, payload.Row, nil, "failed", err.Error())
		return resultRow
	}

	resultRow.Status = "success"
	if !isNew {
		resultRow.Message = "data diperbarui (sudah ada sebelumnya)"
	}
	s.logDetail(logID, payload.Page, payload.Row, &siswaID, "success", resultRow.Message)
	return resultRow
}

func (s *ImportService) logDetail(logID int64, page, row int, siswaID *int64, status, errMsg string) {
	detail := &domain.ImportLogDetail{
		ImportLogID: logID,
		SourcePage:  &page,
		SourceRow:   &row,
		SiswaID:     siswaID,
		Status:      status,
	}
	if errMsg != "" {
		detail.ErrorMessage = &errMsg
	}

	if err := s.importLogRepo.AddDetail(detail); err != nil {
		log.Printf("warning: gagal simpan detail log: %v", err)
	}
}

// hasContent — cek apakah map punya minimal 1 field terisi (bukan semua kosong/"-")
func hasContent(raw map[string]string) bool {
	for _, v := range raw {
		if v != "" && v != "-" {
			return true
		}
	}
	return false
}

func (s *ImportService) GetJobStatus(jobID string) (*domain.ImportLog, error) {
	jobLog, err := s.importLogRepo.GetJobByID(jobID)
	if err != nil {
		return nil, err
	}
	if jobLog == nil {
		return nil, fmt.Errorf("job tidak ditemukan")
	}
	return jobLog, nil
}
