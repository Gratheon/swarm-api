-- +goose Up
CREATE TABLE IF NOT EXISTS hive_logs (
  id INT NOT NULL AUTO_INCREMENT,
  user_id INT UNSIGNED NOT NULL,
  hive_id INT UNSIGNED NOT NULL,
  action VARCHAR(64) NOT NULL,
  title VARCHAR(255) NOT NULL,
  details TEXT NULL,
  source VARCHAR(32) NULL,
  related_hives JSON NULL,
  dedupe_key VARCHAR(255) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  active TINYINT(1) NOT NULL DEFAULT 1,
  PRIMARY KEY (id),
  KEY idx_hive_logs_hive_created (user_id, hive_id, created_at),
  UNIQUE KEY uniq_hive_logs_user_dedupe (user_id, dedupe_key),
  CONSTRAINT fk_hive_logs_hive FOREIGN KEY (hive_id) REFERENCES hives(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +goose Down
DROP TABLE IF EXISTS hive_logs;
