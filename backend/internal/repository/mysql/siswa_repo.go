package mysql

import (
	"database/sql"
	"errors"

	"bank-data/backend/internal/domain"

	"github.com/google/uuid"

	"bank-data/backend/internal/repository"
)

type SiswaRepo struct {
	db *sql.DB
}

func NewSiswaRepo(db *sql.DB) *SiswaRepo {
	return &SiswaRepo{db: db}
}

// FindByNIK — dipakai untuk cek duplikat sebelum insert/upsert
func (r *SiswaRepo) FindByNIK(nik string) (*domain.Siswa, error) {
	if nik == "" {
		return nil, nil
	}

	query := `SELECT id, uuid, nik, nisn, nama, source
	          FROM siswa WHERE nik = ? AND deleted_at IS NULL LIMIT 1`

	var s domain.Siswa
	err := r.db.QueryRow(query, nik).Scan(&s.ID, &s.UUID, &s.NIK, &s.NISN, &s.Nama, &s.Source)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // tidak ketemu, bukan error
	}
	if err != nil {
		return nil, err
	}

	return &s, nil
}

// Insert — selalu buat record baru, dipakai untuk mode "insert"
func (r *SiswaRepo) Insert(
	siswa *domain.Siswa,
	wali []domain.WaliSiswa,
	aktivitas []domain.AktivitasBelajar,
	beasiswa []domain.Beasiswa,
	prestasi []domain.Prestasi,
) (int64, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() // no-op kalau sudah di-commit

	siswaID, err := r.insertSiswaTx(tx, siswa)
	if err != nil {
		return 0, err
	}

	if err := r.insertRelasiTx(tx, siswaID, wali, aktivitas, beasiswa, prestasi); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return siswaID, nil
}

// Upsert — update kalau NIK sudah ada, insert kalau belum
// return: (siswaID, isNewRecord, error)
func (r *SiswaRepo) Upsert(
	siswa *domain.Siswa,
	wali []domain.WaliSiswa,
	aktivitas []domain.AktivitasBelajar,
	beasiswa []domain.Beasiswa,
	prestasi []domain.Prestasi,
) (int64, bool, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, false, err
	}
	defer tx.Rollback()

	var existingID int64
	var isNew bool

	if siswa.NIK != nil && *siswa.NIK != "" {
		err := tx.QueryRow(`SELECT id FROM siswa WHERE nik = ? AND deleted_at IS NULL LIMIT 1`, *siswa.NIK).Scan(&existingID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return 0, false, err
		}
	}

	if existingID == 0 {
		// belum ada -> insert baru
		isNew = true
		existingID, err = r.insertSiswaTx(tx, siswa)
		if err != nil {
			return 0, false, err
		}
	} else {
		// sudah ada -> update
		isNew = false
		if err := r.updateSiswaTx(tx, existingID, siswa); err != nil {
			return 0, false, err
		}

		// hapus relasi lama sebelum insert ulang (lebih simpel daripada diff per-field)
		if err := r.deleteRelasiTx(tx, existingID); err != nil {
			return 0, false, err
		}
	}

	if err := r.insertRelasiTx(tx, existingID, wali, aktivitas, beasiswa, prestasi); err != nil {
		return 0, false, err
	}

	if err := tx.Commit(); err != nil {
		return 0, false, err
	}

	return existingID, isNew, nil
}

// ===== helper internal (dipakai dalam transaction) =====

func (r *SiswaRepo) insertSiswaTx(tx *sql.Tx, s *domain.Siswa) (int64, error) {
	newUUID := uuid.New().String()

	query := `INSERT INTO siswa (
		uuid, nik, nisn, kip, nama, tempat_lahir, tanggal_lahir, jenis_kelamin, agama,
		jumlah_saudara, anak_ke, hobi, cita_cita, no_handphone, email, yang_membiayai,
		kebutuhan_disabilitas, kebutuhan_khusus, alamat, status_tempat_tinggal,
		jarak_rumah_madrasah, waktu_tempuh, transportasi, fingerprint, rfid,
		source, source_page, source_row, is_failed, raw_json
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := tx.Exec(query,
		newUUID, s.NIK, s.NISN, s.KIP, s.Nama, s.TempatLahir, s.TanggalLahir, s.JenisKelamin, s.Agama,
		s.JumlahSaudara, s.AnakKe, s.Hobi, s.CitaCita, s.NoHandphone, s.Email, s.YangMembiayai,
		s.KebutuhanDisabilitas, s.KebutuhanKhusus, s.Alamat, s.StatusTempatTinggal,
		s.JarakRumahMadrasah, s.WaktuTempuh, s.Transportasi, s.Fingerprint, s.RFID,
		s.Source, s.SourcePage, s.SourceRow, s.IsFailed, s.RawJSON,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *SiswaRepo) updateSiswaTx(tx *sql.Tx, id int64, s *domain.Siswa) error {
	query := `UPDATE siswa SET
		nisn=?, kip=?, nama=?, tempat_lahir=?, tanggal_lahir=?, jenis_kelamin=?, agama=?,
		jumlah_saudara=?, anak_ke=?, hobi=?, cita_cita=?, no_handphone=?, email=?, yang_membiayai=?,
		kebutuhan_disabilitas=?, kebutuhan_khusus=?, alamat=?, status_tempat_tinggal=?,
		jarak_rumah_madrasah=?, waktu_tempuh=?, transportasi=?, fingerprint=?, rfid=?,
		source=?, source_page=?, source_row=?, is_failed=?, raw_json=?, updated_at=NOW()
	WHERE id=?`

	_, err := tx.Exec(query,
		s.NISN, s.KIP, s.Nama, s.TempatLahir, s.TanggalLahir, s.JenisKelamin, s.Agama,
		s.JumlahSaudara, s.AnakKe, s.Hobi, s.CitaCita, s.NoHandphone, s.Email, s.YangMembiayai,
		s.KebutuhanDisabilitas, s.KebutuhanKhusus, s.Alamat, s.StatusTempatTinggal,
		s.JarakRumahMadrasah, s.WaktuTempuh, s.Transportasi, s.Fingerprint, s.RFID,
		s.Source, s.SourcePage, s.SourceRow, s.IsFailed, s.RawJSON,
		id,
	)
	return err
}

// deleteRelasiTx — hapus semua relasi lama (wali, aktivitas, beasiswa, prestasi)
// dipanggil sebelum insert ulang saat upsert, supaya tidak dobel/basi
func (r *SiswaRepo) deleteRelasiTx(tx *sql.Tx, siswaID int64) error {
	tables := []string{"wali_siswa", "aktivitas_belajar", "beasiswa", "prestasi"}
	for _, table := range tables {
		if _, err := tx.Exec("DELETE FROM "+table+" WHERE siswa_id = ?", siswaID); err != nil {
			return err
		}
	}
	return nil
}

func (r *SiswaRepo) insertRelasiTx(
	tx *sql.Tx,
	siswaID int64,
	wali []domain.WaliSiswa,
	aktivitas []domain.AktivitasBelajar,
	beasiswa []domain.Beasiswa,
	prestasi []domain.Prestasi,
) error {
	// wali (ayah, ibu, wali)
	for _, w := range wali {
		_, err := tx.Exec(`INSERT INTO wali_siswa (
			siswa_id, role, nik, nama_lengkap, tempat_lahir, tanggal_lahir, status,
			pendidikan_terakhir, pekerjaan_utama, domisili, no_handphone, penghasilan,
			alamat, status_tempat_tinggal
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			siswaID, w.Role, w.NIK, w.NamaLengkap, w.TempatLahir, w.TanggalLahir, w.Status,
			w.PendidikanTerakhir, w.PekerjaanUtama, w.Domisili, w.NoHandphone, w.Penghasilan,
			w.Alamat, w.StatusTempatTinggal,
		)
		if err != nil {
			return err
		}
	}

	// aktivitas belajar — skip kalau semua field kosong (seperti di contoh JSON awal)
	for _, a := range aktivitas {
		if isAktivitasKosong(a) {
			continue
		}
		_, err := tx.Exec(`INSERT INTO aktivitas_belajar (
			siswa_id, tahun_ajaran_semester, tanggal_mulai_masuk, nsm_nama_lembaga,
			tingkat_kelompok, jurusan, status_keaktifan, keterangan
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			siswaID, a.TahunAjaranSemester, a.TanggalMulaiMasuk, a.NSMNamaLembaga,
			a.TingkatKelompok, a.Jurusan, a.StatusKeaktifan, a.Keterangan,
		)
		if err != nil {
			return err
		}
	}

	// beasiswa
	for _, b := range beasiswa {
		_, err := tx.Exec(`INSERT INTO beasiswa (siswa_id, nama_beasiswa, tahun, keterangan) VALUES (?, ?, ?, ?)`,
			siswaID, b.NamaBeasiswa, b.Tahun, b.Keterangan,
		)
		if err != nil {
			return err
		}
	}

	// prestasi
	for _, p := range prestasi {
		_, err := tx.Exec(`INSERT INTO prestasi (siswa_id, nama_prestasi, tingkat, tahun, keterangan) VALUES (?, ?, ?, ?, ?)`,
			siswaID, p.NamaPrestasi, p.Tingkat, p.Tahun, p.Keterangan,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func isAktivitasKosong(a domain.AktivitasBelajar) bool {
	return (a.TahunAjaranSemester == nil || *a.TahunAjaranSemester == "") &&
		(a.NSMNamaLembaga == nil || *a.NSMNamaLembaga == "")
}

// ===== Read operations =====

func (r *SiswaRepo) FindByID(id int64) (*domain.Siswa, error) {
	query := `SELECT id, uuid, nik, nisn, kip, nama, tempat_lahir, tanggal_lahir, jenis_kelamin,
		agama, jumlah_saudara, anak_ke, hobi, cita_cita, no_handphone, email, yang_membiayai,
		kebutuhan_disabilitas, kebutuhan_khusus, alamat, status_tempat_tinggal,
		jarak_rumah_madrasah, waktu_tempuh, transportasi, fingerprint, rfid,
		source, created_at, updated_at
	FROM siswa WHERE id = ? AND deleted_at IS NULL`

	var s domain.Siswa
	err := r.db.QueryRow(query, id).Scan(
		&s.ID, &s.UUID, &s.NIK, &s.NISN, &s.KIP, &s.Nama, &s.TempatLahir, &s.TanggalLahir, &s.JenisKelamin,
		&s.Agama, &s.JumlahSaudara, &s.AnakKe, &s.Hobi, &s.CitaCita, &s.NoHandphone, &s.Email, &s.YangMembiayai,
		&s.KebutuhanDisabilitas, &s.KebutuhanKhusus, &s.Alamat, &s.StatusTempatTinggal,
		&s.JarakRumahMadrasah, &s.WaktuTempuh, &s.Transportasi, &s.Fingerprint, &s.RFID,
		&s.Source, &s.CreatedAt, &s.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (r *SiswaRepo) List(limit, offset int, search string) ([]domain.Siswa, int, error) {
	var siswas []domain.Siswa
	var total int

	countQuery := `SELECT COUNT(*) FROM siswa WHERE deleted_at IS NULL`
	listQuery := `SELECT id, uuid, nik, nisn, nama, jenis_kelamin, source, created_at
	              FROM siswa WHERE deleted_at IS NULL`

	args := []interface{}{}
	if search != "" {
		countQuery += ` AND (nama LIKE ? OR nik LIKE ? OR nisn LIKE ?)`
		listQuery += ` AND (nama LIKE ? OR nik LIKE ? OR nisn LIKE ?)`
		searchTerm := "%" + search + "%"
		args = append(args, searchTerm, searchTerm, searchTerm)
	}

	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listQuery += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := r.db.Query(listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var s domain.Siswa
		if err := rows.Scan(&s.ID, &s.UUID, &s.NIK, &s.NISN, &s.Nama, &s.JenisKelamin, &s.Source, &s.CreatedAt); err != nil {
			return nil, 0, err
		}
		siswas = append(siswas, s)
	}

	return siswas, total, nil
}

func (r *SiswaRepo) Delete(id int64) error {
	_, err := r.db.Exec(`UPDATE siswa SET deleted_at = NOW() WHERE id = ?`, id)
	return err
}

// UpdateBiodataOnly - khusus edit biodata manual, tidak pernah hapus/ubah relasi
func (r *SiswaRepo) UpdateBiodataOnly(id int64, s *domain.Siswa) error {
	query := `UPDATE siswa SET
		nisn=?, kip=?, nama=?, tempat_lahir=?, tanggal_lahir=?, jenis_kelamin=?, agama=?,
		jumlah_saudara=?, anak_ke=?, hobi=?, cita_cita=?, no_handphone=?, email=?, yang_membiayai=?,
		kebutuhan_disabilitas=?, kebutuhan_khusus=?, alamat=?, status_tempat_tinggal=?,
		jarak_rumah_madrasah=?, waktu_tempuh=?, transportasi=?, fingerprint=?, rfid=?,
		updated_at=NOW()
	WHERE id=? AND deleted_at IS NULL`

	result, err := r.db.Exec(query,
		s.NISN, s.KIP, s.Nama, s.TempatLahir, s.TanggalLahir, s.JenisKelamin, s.Agama,
		s.JumlahSaudara, s.AnakKe, s.Hobi, s.CitaCita, s.NoHandphone, s.Email, s.YangMembiayai,
		s.KebutuhanDisabilitas, s.KebutuhanKhusus, s.Alamat, s.StatusTempatTinggal,
		s.JarakRumahMadrasah, s.WaktuTempuh, s.Transportasi, s.Fingerprint, s.RFID,
		id,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("siswa tidak ditemukan atau sudah dihapus")
	}

	return nil
}

func (r *SiswaRepo) FindAllForExport(filter repository.ExportFilter) ([]domain.SiswaExportData, error) {
	query := `SELECT id, uuid, nik, nisn, kip, nama, tempat_lahir, tanggal_lahir, jenis_kelamin,
		agama, jumlah_saudara, anak_ke, hobi, cita_cita, no_handphone, email, yang_membiayai,
		kebutuhan_disabilitas, kebutuhan_khusus, alamat, status_tempat_tinggal,
		jarak_rumah_madrasah, waktu_tempuh, transportasi, fingerprint, rfid,
		source, created_at, updated_at
	FROM siswa WHERE deleted_at IS NULL`

	args := []interface{}{}

	if filter.Search != "" {
		query += ` AND (nama LIKE ? OR nik LIKE ? OR nisn LIKE ?)`
		term := "%" + filter.Search + "%"
		args = append(args, term, term, term)
	}
	if filter.Source != "" {
		query += ` AND source = ?`
		args = append(args, filter.Source)
	}
	if filter.DateFrom != "" {
		query += ` AND created_at >= ?`
		args = append(args, filter.DateFrom)
	}
	if filter.DateTo != "" {
		query += ` AND created_at <= ?`
		args = append(args, filter.DateTo)
	}

	query += ` ORDER BY nama ASC`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.SiswaExportData

	for rows.Next() {
		var s domain.Siswa
		err := rows.Scan(
			&s.ID, &s.UUID, &s.NIK, &s.NISN, &s.KIP, &s.Nama, &s.TempatLahir, &s.TanggalLahir, &s.JenisKelamin,
			&s.Agama, &s.JumlahSaudara, &s.AnakKe, &s.Hobi, &s.CitaCita, &s.NoHandphone, &s.Email, &s.YangMembiayai,
			&s.KebutuhanDisabilitas, &s.KebutuhanKhusus, &s.Alamat, &s.StatusTempatTinggal,
			&s.JarakRumahMadrasah, &s.WaktuTempuh, &s.Transportasi, &s.Fingerprint, &s.RFID,
			&s.Source, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		exportData := domain.SiswaExportData{Siswa: s}
		results = append(results, exportData)
	}

	// ambil relasi untuk tiap siswa (N+1 query - oke untuk skala ratusan/ribuan, perlu dioptimasi kalau jutaan)
	for i := range results {
		siswaID := results[i].Siswa.ID

		waliRows, err := r.db.Query(`SELECT role, nik, nama_lengkap, tempat_lahir, tanggal_lahir, status,
			pendidikan_terakhir, pekerjaan_utama, domisili, no_handphone, penghasilan, alamat, status_tempat_tinggal
			FROM wali_siswa WHERE siswa_id = ?`, siswaID)
		if err != nil {
			return nil, err
		}

		for waliRows.Next() {
			var w domain.WaliSiswa
			if err := waliRows.Scan(&w.Role, &w.NIK, &w.NamaLengkap, &w.TempatLahir, &w.TanggalLahir, &w.Status,
				&w.PendidikanTerakhir, &w.PekerjaanUtama, &w.Domisili, &w.NoHandphone, &w.Penghasilan,
				&w.Alamat, &w.StatusTempatTinggal); err != nil {
				waliRows.Close()
				return nil, err
			}

			switch w.Role {
			case "ayah":
				wCopy := w
				results[i].Ayah = &wCopy
			case "ibu":
				wCopy := w
				results[i].Ibu = &wCopy
			case "wali":
				wCopy := w
				results[i].Wali = &wCopy
			}
		}
		waliRows.Close()

		aktRows, err := r.db.Query(`SELECT tahun_ajaran_semester, tanggal_mulai_masuk, nsm_nama_lembaga,
			tingkat_kelompok, jurusan, status_keaktifan, keterangan
			FROM aktivitas_belajar WHERE siswa_id = ?`, siswaID)
		if err != nil {
			return nil, err
		}
		for aktRows.Next() {
			var a domain.AktivitasBelajar
			if err := aktRows.Scan(&a.TahunAjaranSemester, &a.TanggalMulaiMasuk, &a.NSMNamaLembaga,
				&a.TingkatKelompok, &a.Jurusan, &a.StatusKeaktifan, &a.Keterangan); err != nil {
				aktRows.Close()
				return nil, err
			}
			results[i].AktivitasBelajar = append(results[i].AktivitasBelajar, a)
		}
		aktRows.Close()
	}

	return results, nil
}

// FindByField - pencarian exact match by rfid, nisn, atau nik
// field harus whitelist - JANGAN terima field sembarangan dari user (celah SQL injection kalau field langsung dari input)
func (r *SiswaRepo) FindByField(field, value string) (*domain.Siswa, error) {
	allowedFields := map[string]bool{
		"rfid": true,
		"nisn": true,
		"nik":  true,
	}

	if !allowedFields[field] {
		return nil, errors.New("field pencarian tidak diizinkan")
	}

	query := `SELECT id, uuid, nik, nisn, kip, nama, tempat_lahir, tanggal_lahir, jenis_kelamin,
		agama, jumlah_saudara, anak_ke, hobi, cita_cita, no_handphone, email, yang_membiayai,
		kebutuhan_disabilitas, kebutuhan_khusus, alamat, status_tempat_tinggal,
		jarak_rumah_madrasah, waktu_tempuh, transportasi, fingerprint, rfid,
		source, created_at, updated_at
	FROM siswa WHERE ` + field + ` = ? AND deleted_at IS NULL LIMIT 1`

	var s domain.Siswa
	err := r.db.QueryRow(query, value).Scan(
		&s.ID, &s.UUID, &s.NIK, &s.NISN, &s.KIP, &s.Nama, &s.TempatLahir, &s.TanggalLahir, &s.JenisKelamin,
		&s.Agama, &s.JumlahSaudara, &s.AnakKe, &s.Hobi, &s.CitaCita, &s.NoHandphone, &s.Email, &s.YangMembiayai,
		&s.KebutuhanDisabilitas, &s.KebutuhanKhusus, &s.Alamat, &s.StatusTempatTinggal,
		&s.JarakRumahMadrasah, &s.WaktuTempuh, &s.Transportasi, &s.Fingerprint, &s.RFID,
		&s.Source, &s.CreatedAt, &s.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &s, nil
}
