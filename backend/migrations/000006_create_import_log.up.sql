CREATE TABLE import_log (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    job_id CHAR(36) NOT NULL UNIQUE,
    source ENUM('pendaftaran','emis_scraping','manual_upload') NOT NULL,
    mode ENUM('insert','upsert') NOT NULL,
    status ENUM('pending','processing','completed','failed') NOT NULL DEFAULT 'pending',
    total_rows INT UNSIGNED DEFAULT 0,
    success_rows INT UNSIGNED DEFAULT 0,
    failed_rows INT UNSIGNED DEFAULT 0,
    duplicate_rows INT UNSIGNED DEFAULT 0,
    triggered_by VARCHAR(100) NULL,
    started_at TIMESTAMP NULL,
    finished_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_job_id (job_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE import_log_detail (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    import_log_id BIGINT UNSIGNED NOT NULL,
    source_page INT NULL,
    source_row INT NULL,
    siswa_id BIGINT UNSIGNED NULL,
    status ENUM('success','failed','duplicate') NOT NULL,
    error_message TEXT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (import_log_id) REFERENCES import_log(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
