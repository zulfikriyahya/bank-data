package service

import (
	"errors"

	"bank-data/backend/internal/repository"
	"bank-data/backend/pkg/jwtutil"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	adminRepo repository.AdminRepository
}

func NewAuthService(adminRepo repository.AdminRepository) *AuthService {
	return &AuthService{adminRepo: adminRepo}
}

func (s *AuthService) Login(username, password string) (string, error) {
	user, err := s.adminRepo.FindByUsername(username)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("username atau password salah")
	}
	if !user.IsActive {
		return "", errors.New("akun tidak aktif")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("username atau password salah")
	}

	token, err := jwtutil.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return "", err
	}

	_ = s.adminRepo.UpdateLastLogin(user.ID)

	return token, nil
}
