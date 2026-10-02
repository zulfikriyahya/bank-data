CREATE TABLE wali_siswa (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    siswa_id BIGINT UNSIGNED NOT NULL,
    role ENUM('ayah','ibu','wali') NOT NULL,
    nik VARCHAR(20) NULL,
    nama_lengkap VARCHAR(150) NULL,
    tempat_lahir VARCHAR(100) NULL,
    tanggal_lahir DATE NULL,
    status VARCHAR(50) NULL,
    pendidikan_terakhir VARCHAR(30) NULL,
    pekerjaan_utama VARCHAR(100) NULL,
    domisili VARCHAR(50) NULL,
    no_handphone VARCHAR(20) NULL,
    penghasilan VARCHAR(50) NULL,
    alamat TEXT NULL,
    status_tempat_tinggal VARCHAR(50) NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (siswa_id) REFERENCES siswa(id) ON DELETE CASCADE,
    UNIQUE KEY uq_siswa_role (siswa_id, role)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
