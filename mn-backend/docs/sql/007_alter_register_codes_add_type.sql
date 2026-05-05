ALTER TABLE register_codes
    ADD COLUMN type VARCHAR(32) NOT NULL DEFAULT 'register' AFTER email,
    DROP PRIMARY KEY,
    ADD PRIMARY KEY (email, type);
