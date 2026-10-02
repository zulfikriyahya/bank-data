package service

import (
	"errors"

	"bank-data/backend/internal/domain"
	"bank-data/backend/internal/repository"
)

type SiswaService struct {
	repo repository.SiswaRepository
}

func NewSiswaService(repo repository.SiswaRepository) *SiswaService {
	return &SiswaService{repo: repo}
}

func (s *SiswaService) GetByID(id int64) (*domain.Siswa, error) {
	siswa, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if siswa == nil {
		return nil, errors.New("siswa tidak ditemukan")
	}
	return siswa, nil
}

func (s *SiswaService) List(page, perPage int, search string) ([]domain.Siswa, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	return s.repo.List(perPage, offset, search)
}

func (s *SiswaService) Update(id int64, siswa *domain.Siswa) error {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("siswa tidak ditemukan")
	}

	// PENTING: pakai UpdateBiodataOnly, BUKAN Upsert
	// supaya relasi wali/aktivitas/beasiswa/prestasi tidak ikut terhapus
	return s.repo.UpdateBiodataOnly(id, siswa)
}

func (s *SiswaService) Delete(id int64) error {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("siswa tidak ditemukan")
	}
	return s.repo.Delete(id)
}

func (s *SiswaService) FindByNIK(nik string) (*domain.Siswa, error) {
	return s.repo.FindByNIK(nik)
}

func (s *SiswaService) FindByField(field, value string) (*domain.Siswa, error) {
	siswa, err := s.repo.FindByField(field, value)
	if err != nil {
		return nil, err
	}
	if siswa == nil {
		return nil, errors.New("siswa tidak ditemukan")
	}
	return siswa, nil
}
