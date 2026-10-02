package domain

import "time"

type ScraperConfig struct {
	ID           int64     `json:"id"`
	ScraperName  string    `json:"scraper_name"`
	Config       map[string]string `json:"config"`
	UpdatedBy    string    `json:"updated_by,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type SetScraperConfigInput struct {
	Config map[string]string `json:"config"`
}
