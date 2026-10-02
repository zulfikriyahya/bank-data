package domain

import "time"

type APIClient struct {
	ID           int64     `json:"id"`
	ClientName   string    `json:"client_name"`
	APIKeyHash   string    `json:"-"` // tidak pernah diekspos lewat API
	Scopes       []string  `json:"scopes"`
	RateLimit    int       `json:"rate_limit"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateAPIClientInput struct {
	ClientName string   `json:"client_name"`
	Scopes     []string `json:"scopes"`
	RateLimit  int      `json:"rate_limit"`
}

type UpdateAPIClientInput struct {
	Scopes    []string `json:"scopes"`
	RateLimit int      `json:"rate_limit"`
	IsActive  *bool    `json:"is_active"`
}
