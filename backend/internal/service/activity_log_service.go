package service

import (
	"encoding/json"
	"errors"

	"bank-data/backend/internal/domain"
	"bank-data/backend/internal/repository"
)

var allowedActivityTypes = map[string]bool{
	"presensi_masuk":    true,
	"presensi_pulang":   true,
	"pinjam_buku":       true,
	"kembalikan_buku":   true,
}

type ActivityLogService struct {
	repo      repository.ActivityLogRepository
	siswaRepo repository.SiswaRepository
}

func NewActivityLogService(repo repository.ActivityLogRepository, siswaRepo repository.SiswaRepository) *ActivityLogService {
	return &ActivityLogService{repo: repo, siswaRepo: siswaRepo}
}

func (s *ActivityLogService) Record(clientID int64, input domain.ActivityLogInput) (int64, error) {
	if !allowedActivityTypes[input.ActivityType] {
		return 0, errors.New("activity_type tidak dikenali, gunakan salah satu: presensi_masuk, presensi_pulang, pinjam_buku, kembalikan_buku")
	}

	// pastikan siswa benar-benar ada, supaya tidak ada log nyangkut ke ID siswa yang tidak valid
	siswa, err := s.siswaRepo.FindByID(input.SiswaID)
	if err != nil {
		return 0, err
	}
	if siswa == nil {
		return 0, errors.New("siswa tidak ditemukan")
	}

	var metadataStr *string
	if input.Metadata != nil {
		metaBytes, err := json.Marshal(input.Metadata)
		if err != nil {
			return 0, errors.New("format metadata tidak valid")
		}
		m := string(metaBytes)
		metadataStr = &m
	}

	var description *string
	if input.Description != "" {
		description = &input.Description
	}

	log := &domain.ActivityLog{
		SiswaID:      input.SiswaID,
		ClientID:     clientID,
		ActivityType: input.ActivityType,
		Description:  description,
		Metadata:     metadataStr,
	}

	return s.repo.Create(log)
}

func (s *ActivityLogService) GetBySiswa(siswaID int64, limit int) ([]domain.ActivityLog, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.ListBySiswa(siswaID, limit)
}

func (s *ActivityLogService) List(page, perPage int, activityType string) ([]domain.ActivityLog, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	return s.repo.List(perPage, offset, activityType)
}
