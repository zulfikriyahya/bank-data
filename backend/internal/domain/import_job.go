package domain

import "time"

type ImportLog struct {
	ID            int64      `json:"id"`
	JobID         string     `json:"job_id"`
	Source        string     `json:"source"` // pendaftaran, emis_scraping, manual_upload
	Mode          string     `json:"mode"`   // insert, upsert
	Status        string     `json:"status"` // pending, processing, completed, failed
	TotalRows     int        `json:"total_rows"`
	SuccessRows   int        `json:"success_rows"`
	FailedRows    int        `json:"failed_rows"`
	DuplicateRows int        `json:"duplicate_rows"`
	TriggeredBy   string     `json:"triggered_by"`
	StartedAt     *time.Time `json:"started_at"`
	FinishedAt    *time.Time `json:"finished_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

type ImportLogDetail struct {
	ID           int64   `json:"id"`
	ImportLogID  int64   `json:"import_log_id"`
	SourcePage   *int    `json:"source_page"`
	SourceRow    *int    `json:"source_row"`
	SiswaID      *int64  `json:"siswa_id"`
	Status       string  `json:"status"` // success, failed, duplicate
	ErrorMessage *string `json:"error_message"`
}

// Hasil ringkas setelah proses import selesai
type ImportResult struct {
	JobID         string             `json:"job_id"`
	TotalRows     int                `json:"total_rows"`
	SuccessRows   int                `json:"success_rows"`
	FailedRows    int                `json:"failed_rows"`
	DuplicateRows int                `json:"duplicate_rows"`
	Details       []ImportResultRow  `json:"details,omitempty"`
}

type ImportResultRow struct {
	Row     int    `json:"row"`
	Page    int    `json:"page"`
	Status  string `json:"status"` // success, failed, duplicate
	Message string `json:"message,omitempty"`
}
