package mysql

import (
	"database/sql"

	"bank-data/backend/internal/domain"
)

type ActivityLogRepo struct {
	db *sql.DB
}

func NewActivityLogRepo(db *sql.DB) *ActivityLogRepo {
	return &ActivityLogRepo{db: db}
}

func (r *ActivityLogRepo) Create(log *domain.ActivityLog) (int64, error) {
	query := `INSERT INTO activity_log (siswa_id, client_id, activity_type, description, metadata)
	          VALUES (?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query, log.SiswaID, log.ClientID, log.ActivityType, log.Description, log.Metadata)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *ActivityLogRepo) ListBySiswa(siswaID int64, limit int) ([]domain.ActivityLog, error) {
	query := `SELECT al.id, al.siswa_id, al.client_id, c.client_name, al.activity_type, 
	          al.description, al.metadata, al.created_at
	          FROM activity_log al
	          JOIN api_clients c ON c.id = al.client_id
	          WHERE al.siswa_id = ?
	          ORDER BY al.created_at DESC
	          LIMIT ?`

	rows, err := r.db.Query(query, siswaID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []domain.ActivityLog
	for rows.Next() {
		var l domain.ActivityLog
		if err := rows.Scan(&l.ID, &l.SiswaID, &l.ClientID, &l.ClientName, &l.ActivityType,
			&l.Description, &l.Metadata, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}

	return logs, nil
}

func (r *ActivityLogRepo) List(limit, offset int, activityType string) ([]domain.ActivityLog, int, error) {
	countQuery := `SELECT COUNT(*) FROM activity_log`
	listQuery := `SELECT al.id, al.siswa_id, al.client_id, c.client_name, al.activity_type,
	              al.description, al.metadata, al.created_at
	              FROM activity_log al
	              JOIN api_clients c ON c.id = al.client_id`

	args := []interface{}{}
	if activityType != "" {
		countQuery += ` WHERE activity_type = ?`
		listQuery += ` WHERE al.activity_type = ?`
		args = append(args, activityType)
	}

	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listQuery += ` ORDER BY al.created_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := r.db.Query(listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []domain.ActivityLog
	for rows.Next() {
		var l domain.ActivityLog
		if err := rows.Scan(&l.ID, &l.SiswaID, &l.ClientID, &l.ClientName, &l.ActivityType,
			&l.Description, &l.Metadata, &l.CreatedAt); err != nil {
			return nil, 0, err
		}
		logs = append(logs, l)
	}

	return logs, total, nil
}
