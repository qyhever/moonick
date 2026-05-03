CREATE TABLE IF NOT EXISTS register_codes (
    email VARCHAR(128) NOT NULL PRIMARY KEY,
    code VARCHAR(6) NOT NULL,
    expires_at DATETIME NOT NULL,
    last_sent_at DATETIME NULL DEFAULT NULL,
    send_window_started_at DATETIME NULL DEFAULT NULL,
    send_count_in_window INT NOT NULL DEFAULT 1,
    used_at DATETIME NULL DEFAULT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    KEY idx_register_codes_expires_at (expires_at)
);
