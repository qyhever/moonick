ALTER TABLE register_codes
    ADD COLUMN last_sent_at DATETIME NULL DEFAULT NULL AFTER expires_at,
    ADD COLUMN send_window_started_at DATETIME NULL DEFAULT NULL AFTER last_sent_at,
    ADD COLUMN send_count_in_window INT NOT NULL DEFAULT 1 AFTER send_window_started_at;
