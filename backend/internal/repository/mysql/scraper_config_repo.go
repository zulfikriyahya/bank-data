package mysql

import (
	"database/sql"
	"errors"

	"bank-data/backend/internal/domain"
)

type ScraperConfigRepo struct {
	db *sql.DB
}

func NewScraperConfigRepo(db *sql.DB) *ScraperConfigRepo {
	return &ScraperConfigRepo{db: db}
}

func (r *ScraperConfigRepo) Upsert(scraperName string, configEncrypted string, updatedBy string) error {
	query := `INSERT INTO scraper_configs (scraper_name, config_encrypted, updated_by)
	          VALUES (?, ?, ?)
	          ON DUPLICATE KEY UPDATE config_encrypted = VALUES(config_encrypted), updated_by = VALUES(updated_by), updated_at = NOW()`

	_, err := r.db.Exec(query, scraperName, configEncrypted, updatedBy)
	return err
}

func (r *ScraperConfigRepo) FindByName(scraperName string) (*domain.ScraperConfig, string, error) {
	query := `SELECT id, scraper_name, config_encrypted, updated_by, created_at, updated_at
	          FROM scraper_configs WHERE scraper_name = ?`

	var c domain.ScraperConfig
	var encrypted string
	var updatedBy sql.NullString

	err := r.db.QueryRow(query, scraperName).Scan(&c.ID, &c.ScraperName, &encrypted, &updatedBy, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, "", nil
	}
	if err != nil {
		return nil, "", err
	}

	c.UpdatedBy = updatedBy.String
	return &c, encrypted, nil
}

func (r *ScraperConfigRepo) List() ([]domain.ScraperConfig, error) {
	query := `SELECT id, scraper_name, updated_by, created_at, updated_at FROM scraper_configs ORDER BY scraper_name`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []domain.ScraperConfig
	for rows.Next() {
		var c domain.ScraperConfig
		var updatedBy sql.NullString
		if err := rows.Scan(&c.ID, &c.ScraperName, &updatedBy, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		c.UpdatedBy = updatedBy.String
		configs = append(configs, c)
	}

	return configs, nil
}
