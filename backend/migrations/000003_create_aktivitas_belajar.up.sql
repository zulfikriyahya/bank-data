CREATE TABLE aktivitas_belajar (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    siswa_id BIGINT UNSIGNED NOT NULL,
    tahun_ajaran_semester VARCHAR(30) NULL,
    tanggal_mulai_masuk DATE NULL,
    nsm_nama_lembaga VARCHAR(150) NULL,
    tingkat_kelompok VARCHAR(30) NULL,
    jurusan VARCHAR(100) NULL,
    status_keaktifan VARCHAR(30) NULL,
    keterangan TEXT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (siswa_id) REFERENCES siswa(id) ON DELETE CASCADE,
    INDEX idx_siswa_tahun (siswa_id, tahun_ajaran_semester)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
