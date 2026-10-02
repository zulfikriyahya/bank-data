package mysql

import (
	"database/sql"
	"errors"

	"bank-data/backend/internal/domain"
)

type ImportLogRepo struct {
	db *sql.DB
}

func NewImportLogRepo(db *sql.DB) *ImportLogRepo {
	return &ImportLogRepo{db: db}
}

func (r *ImportLogRepo) CreateJob(log *domain.ImportLog) (int64, error) {
	query := `INSERT INTO import_log (job_id, source, mode, status, total_rows, triggered_by, started_at)
	          VALUES (?, ?, ?, ?, ?, ?, NOW())`

	result, err := r.db.Exec(query, log.JobID, log.Source, log.Mode, log.Status, log.TotalRows, log.TriggeredBy)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *ImportLogRepo) UpdateJobStatus(jobID string, status string, success, failed, duplicate int) error {
	query := `UPDATE import_log
	          SET status=?, success_rows=?, failed_rows=?, duplicate_rows=?, finished_at=NOW()
	          WHERE job_id=?`

	_, err := r.db.Exec(query, status, success, failed, duplicate, jobID)
	return err
}

func (r *ImportLogRepo) AddDetail(detail *domain.ImportLogDetail) error {
	query := `INSERT INTO import_log_detail (import_log_id, source_page, source_row, siswa_id, status, error_message)
	          VALUES (?, ?, ?, ?, ?, ?)`

	_, err := r.db.Exec(query, detail.ImportLogID, detail.SourcePage, detail.SourceRow, detail.SiswaID, detail.Status, detail.ErrorMessage)
	return err
}

func (r *ImportLogRepo) GetJobByID(jobID string) (*domain.ImportLog, error) {
	query := `SELECT id, job_id, source, mode, status, total_rows, success_rows, failed_rows, duplicate_rows, created_at
	          FROM import_log WHERE job_id = ?`

	var log domain.ImportLog
	err := r.db.QueryRow(query, jobID).Scan(
		&log.ID, &log.JobID, &log.Source, &log.Mode, &log.Status,
		&log.TotalRows, &log.SuccessRows, &log.FailedRows, &log.DuplicateRows, &log.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &log, nil
}
