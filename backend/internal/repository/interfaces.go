package repository

import "bank-data/backend/internal/domain"

type SiswaRepository interface {
	FindByNIK(nik string) (*domain.Siswa, error)
	Insert(siswa *domain.Siswa, wali []domain.WaliSiswa, aktivitas []domain.AktivitasBelajar, beasiswa []domain.Beasiswa, prestasi []domain.Prestasi) (int64, error)
	Upsert(siswa *domain.Siswa, wali []domain.WaliSiswa, aktivitas []domain.AktivitasBelajar, beasiswa []domain.Beasiswa, prestasi []domain.Prestasi) (int64, bool, error)

	// UpdateBiodataOnly - update field tabel siswa saja, TIDAK menyentuh wali/aktivitas/beasiswa/prestasi
	// dipakai untuk endpoint PUT /admin/siswa/:id (edit manual lewat dashboard)
	UpdateBiodataOnly(id int64, siswa *domain.Siswa) error

	FindByID(id int64) (*domain.Siswa, error)
	List(limit, offset int, search string) ([]domain.Siswa, int, error)
	Delete(id int64) error
	// FindAllForExport - ambil semua data siswa + relasi lengkap, dengan filter opsional
	FindAllForExport(filter ExportFilter) ([]domain.SiswaExportData, error)
	// FindByField - pencarian spesifik satu field (dipakai consumer app, bukan search umum)
	FindByField(field, value string) (*domain.Siswa, error)
}

type ExportFilter struct {
	Search   string
	Source   string // kosong = semua, atau "pendaftaran"/"emis_scraping"/"manual"
	DateFrom string
	DateTo   string
}

type ImportLogRepository interface {
	CreateJob(log *domain.ImportLog) (int64, error)
	UpdateJobStatus(jobID string, status string, success, failed, duplicate int) error
	AddDetail(detail *domain.ImportLogDetail) error
	GetJobByID(jobID string) (*domain.ImportLog, error)
}

type AdminRepository interface {
	FindByUsername(username string) (*domain.AdminUser, error)
	UpdateLastLogin(id int64) error
}

type ActivityLogRepository interface {
	Create(log *domain.ActivityLog) (int64, error)
	ListBySiswa(siswaID int64, limit int) ([]domain.ActivityLog, error)
	List(limit, offset int, activityType string) ([]domain.ActivityLog, int, error)
}
