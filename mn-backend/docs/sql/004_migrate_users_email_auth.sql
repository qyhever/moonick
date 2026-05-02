-- 将用户账号体系从手机号登录迁移为邮箱登录
-- 适用于已存在 users 表的库
-- 注意：该脚本会为历史用户回填临时邮箱，后续应按真实业务数据修正

START TRANSACTION;

ALTER TABLE users
    ADD COLUMN email VARCHAR(128) NULL AFTER id;

UPDATE users
SET email = CASE
    WHEN COALESCE(TRIM(phone), '') <> '' THEN CONCAT(phone, '@migration.local')
    ELSE CONCAT('legacy_', id, '@migration.local')
END
WHERE email IS NULL OR TRIM(email) = '';

ALTER TABLE users
    MODIFY COLUMN email VARCHAR(128) NOT NULL,
    MODIFY COLUMN phone VARCHAR(32) NOT NULL DEFAULT '';

ALTER TABLE users
    DROP INDEX uk_users_phone,
    ADD UNIQUE KEY uk_users_email (email);

COMMIT;
