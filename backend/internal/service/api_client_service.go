package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"

	"bank-data/backend/internal/domain"
	"bank-data/backend/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type APIClientService struct {
	repo repository.APIClientRepository
}

func NewAPIClientService(repo repository.APIClientRepository) *APIClientService {
	return &APIClientService{repo: repo}
}

// Create - generate API key baru, simpan hash-nya, return KEY ASLI (cuma sekali ini saja ditampilkan)
func (s *APIClientService) Create(input domain.CreateAPIClientInput) (string, *domain.APIClient, error) {
	if input.ClientName == "" {
		return "", nil, errors.New("client_name wajib diisi")
	}
	if len(input.Scopes) == 0 {
		return "", nil, errors.New("scopes wajib diisi minimal 1")
	}

	existing, err := s.repo.FindByName(input.ClientName)
	if err != nil {
		return "", nil, err
	}
	if existing != nil {
		return "", nil, errors.New("client_name sudah terdaftar, gunakan nama lain")
	}

	rateLimit := input.RateLimit
	if rateLimit <= 0 {
		rateLimit = 1000
	}

	apiKey, err := generateAPIKey()
	if err != nil {
		return "", nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(apiKey), bcrypt.DefaultCost)
	if err != nil {
		return "", nil, err
	}

	client := &domain.APIClient{
		ClientName: input.ClientName,
		Scopes:     input.Scopes,
		RateLimit:  rateLimit,
	}

	id, err := s.repo.Create(client, string(hash))
	if err != nil {
		return "", nil, err
	}

	created, err := s.repo.FindByID(id)
	if err != nil {
		return "", nil, err
	}

	return apiKey, created, nil
}

func (s *APIClientService) List() ([]domain.APIClient, error) {
	return s.repo.List()
}

func (s *APIClientService) GetByID(id int64) (*domain.APIClient, error) {
	client, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, errors.New("api client tidak ditemukan")
	}
	return client, nil
}

func (s *APIClientService) Update(id int64, input domain.UpdateAPIClientInput) error {
	return s.repo.Update(id, input)
}

func (s *APIClientService) Delete(id int64) error {
	return s.repo.Delete(id)
}

// RegenerateKey - buat API key baru untuk client yang sudah ada (key lama otomatis tidak valid lagi)
func (s *APIClientService) RegenerateKey(id int64) (string, error) {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return "", err
	}
	if existing == nil {
		return "", errors.New("api client tidak ditemukan")
	}

	apiKey, err := generateAPIKey()
	if err != nil {
		return "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(apiKey), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	// pakai Update tapi perlu query khusus karena interface Update tidak cover ganti hash
	// jadi akses langsung lewat type assertion ke repo mysql - lihat catatan di bawah
	type hashUpdater interface {
		UpdateHash(id int64, hash string) error
	}

	if updater, ok := s.repo.(hashUpdater); ok {
		if err := updater.UpdateHash(id, string(hash)); err != nil {
			return "", err
		}
	} else {
		return "", errors.New("repository tidak mendukung regenerate key")
	}

	return apiKey, nil
}

func generateAPIKey() (string, error) {
	randomBytes := make([]byte, 24)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return "bdk_" + hex.EncodeToString(randomBytes), nil
}
