package mysql

import (
	"database/sql"
	"errors"

	"bank-data/backend/internal/domain"
)

type AdminRepo struct {
	db *sql.DB
}

func NewAdminRepo(db *sql.DB) *AdminRepo {
	return &AdminRepo{db: db}
}

func (r *AdminRepo) FindByUsername(username string) (*domain.AdminUser, error) {
	query := `SELECT id, username, password_hash, role, is_active FROM admin_users WHERE username = ? LIMIT 1`

	var u domain.AdminUser
	err := r.db.QueryRow(query, username).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.IsActive)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *AdminRepo) UpdateLastLogin(id int64) error {
	_, err := r.db.Exec(`UPDATE admin_users SET last_login_at = NOW() WHERE id = ?`, id)
	return err
}
