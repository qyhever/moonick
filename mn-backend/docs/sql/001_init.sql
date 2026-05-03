CREATE TABLE IF NOT EXISTS users (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(128) NOT NULL,
    phone VARCHAR(32) NOT NULL DEFAULT '',
    password_hash VARCHAR(255) NOT NULL,
    nickname VARCHAR(128) NOT NULL,
    avatar_url VARCHAR(512) NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    default_phone VARCHAR(32) NOT NULL DEFAULT '',
    default_wechat VARCHAR(128) NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_users_email (email)
);

CREATE TABLE IF NOT EXISTS admins (
    id BIGINT NOT NULL PRIMARY KEY,
    username VARCHAR(64) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    display_name VARCHAR(128) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_admins_username (username)
);

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

CREATE TABLE IF NOT EXISTS trips (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    publisher_user_id BIGINT NOT NULL,
    trip_type VARCHAR(32) NOT NULL,
    from_text VARCHAR(255) NOT NULL,
    to_text VARCHAR(255) NOT NULL,
    departure_date DATE NOT NULL,
    departure_time TIME NOT NULL,
    seat_count INT NOT NULL DEFAULT 0,
    price_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    is_price_negotiable TINYINT(1) NOT NULL DEFAULT 0,
    contact_wechat VARCHAR(128) NOT NULL DEFAULT '',
    contact_phone VARCHAR(32) NOT NULL DEFAULT '',
    remark TEXT NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    closed_reason VARCHAR(255) NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL DEFAULT NULL,
    KEY idx_trips_publisher_user_id (publisher_user_id),
    KEY idx_trips_status_departure (status, departure_date, departure_time),
    KEY idx_trips_deleted_at (deleted_at)
);

CREATE TABLE IF NOT EXISTS trip_favorites (
    user_id BIGINT NOT NULL,
    trip_id BIGINT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, trip_id),
    KEY idx_trip_favorites_trip_id (trip_id)
);
