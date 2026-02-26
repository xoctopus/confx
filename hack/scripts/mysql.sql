CREATE DATABASE test;

USE test;

CREATE TABLE x (datetime datetime(3), timestamp bigint);

INSERT INTO x (datetime,timestamp) VALUES ('1988-10-24 07:00:00.123', 593650800123);

CREATE TABLE IF NOT EXISTS users (
    f_id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    f_user_id BIGINT UNSIGNED NOT NULL,
    f_name VARCHAR(64) NOT NULL,
    f_email VARCHAR(255) NOT NULL,
    f_token VARBINARY(1024),
    f_balance DECIMAL(22,4) NOT NULL,
    f_created_at BIGINT UNSIGNED NOT NULL,
    f_updated_at BIGINT UNSIGNED NOT NULL,
    f_deleted_at BIGINT UNSIGNED NOT NULL DEFAULT 0,
    UNIQUE KEY ui_email (f_email,f_deleted_at),
    INDEX i_name (f_name,f_deleted_at),
    INDEX i_created_at (f_created_at),
    INDEX i_updated_at (f_updated_at),
    UNIQUE KEY ui_user_id (f_user_id,f_deleted_at),
    PRIMARY KEY (f_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
