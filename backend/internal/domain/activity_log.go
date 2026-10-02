package domain

import "time"

type ActivityLog struct {
	ID           int64     `json:"id"`
	SiswaID      int64     `json:"siswa_id"`
	ClientID     int64     `json:"client_id"`
	ClientName   string    `json:"client_name,omitempty"` // diisi lewat JOIN saat dibaca
	ActivityType string    `json:"activity_type"`
	Description  *string   `json:"description"`
	Metadata     *string   `json:"metadata"` // disimpan sebagai JSON string
	CreatedAt    time.Time `json:"created_at"`
}

// ActivityLogInput - payload yang diterima dari consumer app
type ActivityLogInput struct {
	SiswaID      int64                  `json:"siswa_id"`
	ActivityType string                 `json:"activity_type"`
	Description  string                 `json:"description"`
	Metadata     map[string]interface{} `json:"metadata"`
}
