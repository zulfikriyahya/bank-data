package service

import (
	"encoding/json"
	"errors"

	"bank-data/backend/internal/domain"
	"bank-data/backend/internal/repository"
	"bank-data/backend/pkg/cryptoutil"
)

var maskedFields = map[string]bool{
	"EMIS_PASSWORD": true,
}

type ScraperConfigService struct {
	repo repository.ScraperConfigRepository
}

func NewScraperConfigService(repo repository.ScraperConfigRepository) *ScraperConfigService {
	return &ScraperConfigService{repo: repo}
}

// Set - dipanggil admin, enkripsi seluruh config sebelum simpan
func (s *ScraperConfigService) Set(scraperName string, input domain.SetScraperConfigInput, updatedBy string) error {
	if len(input.Config) == 0 {
		return errors.New("config tidak boleh kosong")
	}

	jsonBytes, err := json.Marshal(input.Config)
	if err != nil {
		return err
	}

	encrypted, err := cryptoutil.Encrypt(string(jsonBytes))
	if err != nil {
		return err
	}

	return s.repo.Upsert(scraperName, encrypted, updatedBy)
}

// GetMasked - dipanggil admin untuk LIHAT config, field sensitif (password) disensor
func (s *ScraperConfigService) GetMasked(scraperName string) (*domain.ScraperConfig, error) {
	config, encrypted, err := s.repo.FindByName(scraperName)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, errors.New("config untuk scraper ini belum pernah diset")
	}

	decrypted, err := cryptoutil.Decrypt(encrypted)
	if err != nil {
		return nil, err
	}

	var configMap map[string]string
	if err := json.Unmarshal([]byte(decrypted), &configMap); err != nil {
		return nil, err
	}

	for key := range configMap {
		if maskedFields[key] {
			configMap[key] = "••••••••"
		}
	}

	config.Config = configMap
	return config, nil
}

// GetFull - HANYA dipanggil oleh scraper sendiri lewat endpoint internal, return apa adanya tanpa masking
func (s *ScraperConfigService) GetFull(scraperName string) (map[string]string, error) {
	_, encrypted, err := s.repo.FindByName(scraperName)
	if err != nil {
		return nil, err
	}
	if encrypted == "" {
		return nil, errors.New("config untuk scraper ini belum pernah diset")
	}

	decrypted, err := cryptoutil.Decrypt(encrypted)
	if err != nil {
		return nil, err
	}

	var configMap map[string]string
	if err := json.Unmarshal([]byte(decrypted), &configMap); err != nil {
		return nil, err
	}

	return configMap, nil
}

func (s *ScraperConfigService) List() ([]domain.ScraperConfig, error) {
	return s.repo.List()
}
