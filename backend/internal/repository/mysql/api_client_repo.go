package mysql

import (
	"database/sql"
	"encoding/json"
	"errors"

	"bank-data/backend/internal/domain"
)

type APIClientRepo struct {
	db *sql.DB
}

func NewAPIClientRepo(db *sql.DB) *APIClientRepo {
	return &APIClientRepo{db: db}
}

func (r *APIClientRepo) Create(client *domain.APIClient, apiKeyHash string) (int64, error) {
	scopesJSON, err := json.Marshal(client.Scopes)
	if err != nil {
		return 0, err
	}

	query := `INSERT INTO api_clients (client_name, api_key_hash, scopes, rate_limit, is_active)
	          VALUES (?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query, client.ClientName, apiKeyHash, string(scopesJSON), client.RateLimit, true)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *APIClientRepo) FindByID(id int64) (*domain.APIClient, error) {
	query := `SELECT id, client_name, scopes, rate_limit, is_active, created_at, updated_at
	          FROM api_clients WHERE id = ?`

	var c domain.APIClient
	var scopesRaw string

	err := r.db.QueryRow(query, id).Scan(&c.ID, &c.ClientName, &scopesRaw, &c.RateLimit, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(scopesRaw), &c.Scopes)
	return &c, nil
}

func (r *APIClientRepo) FindByName(name string) (*domain.APIClient, error) {
	query := `SELECT id, client_name, scopes, rate_limit, is_active, created_at, updated_at
	          FROM api_clients WHERE client_name = ?`

	var c domain.APIClient
	var scopesRaw string

	err := r.db.QueryRow(query, name).Scan(&c.ID, &c.ClientName, &scopesRaw, &c.RateLimit, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(scopesRaw), &c.Scopes)
	return &c, nil
}

func (r *APIClientRepo) List() ([]domain.APIClient, error) {
	query := `SELECT id, client_name, scopes, rate_limit, is_active, created_at, updated_at
	          FROM api_clients ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []domain.APIClient
	for rows.Next() {
		var c domain.APIClient
		var scopesRaw string

		if err := rows.Scan(&c.ID, &c.ClientName, &scopesRaw, &c.RateLimit, &c.IsActive, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(scopesRaw), &c.Scopes)
		clients = append(clients, c)
	}

	return clients, nil
}

func (r *APIClientRepo) Update(id int64, input domain.UpdateAPIClientInput) error {
	existing, err := r.FindByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("api client tidak ditemukan")
	}

	scopes := existing.Scopes
	if input.Scopes != nil {
		scopes = input.Scopes
	}
	scopesJSON, err := json.Marshal(scopes)
	if err != nil {
		return err
	}

	rateLimit := existing.RateLimit
	if input.RateLimit > 0 {
		rateLimit = input.RateLimit
	}

	isActive := existing.IsActive
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	query := `UPDATE api_clients SET scopes=?, rate_limit=?, is_active=?, updated_at=NOW() WHERE id=?`
	_, err = r.db.Exec(query, string(scopesJSON), rateLimit, isActive, id)
	return err
}

func (r *APIClientRepo) Delete(id int64) error {
	result, err := r.db.Exec(`DELETE FROM api_clients WHERE id = ?`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("api client tidak ditemukan")
	}
	return nil
}

func (r *APIClientRepo) UpdateHash(id int64, hash string) error {
	_, err := r.db.Exec(`UPDATE api_clients SET api_key_hash = ?, updated_at = NOW() WHERE id = ?`, hash, id)
	return err
}
