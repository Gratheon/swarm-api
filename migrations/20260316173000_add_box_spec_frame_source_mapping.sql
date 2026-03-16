-- +goose Up
CREATE TABLE IF NOT EXISTS `box_spec_frame_sources` (
  `id` int NOT NULL AUTO_INCREMENT,
  `box_spec_id` int NOT NULL,
  `frame_source_system_id` int NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_box_spec_frame_source` (`box_spec_id`),
  KEY `idx_box_spec_frame_source_system` (`frame_source_system_id`),
  CONSTRAINT `fk_box_spec_frame_source_box_spec` FOREIGN KEY (`box_spec_id`) REFERENCES `box_specs`(`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_box_spec_frame_source_system` FOREIGN KEY (`frame_source_system_id`) REFERENCES `box_systems`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO `box_spec_frame_sources` (`box_spec_id`, `frame_source_system_id`)
SELECT
  bs.id AS box_spec_id,
  COALESCE(src.system_id, bs.system_id) AS frame_source_system_id
FROM `box_specs` bs
LEFT JOIN (
  SELECT
    c.box_spec_id,
    MIN(fs.system_id) AS system_id
  FROM `frame_spec_compatibility` c
  INNER JOIN `frame_specs` fs ON fs.id = c.frame_spec_id
  WHERE fs.active = 1
    AND fs.frame_type = 'FOUNDATION'
  GROUP BY c.box_spec_id
) src ON src.box_spec_id = bs.id
LEFT JOIN `box_spec_frame_sources` existing ON existing.box_spec_id = bs.id
WHERE bs.active = 1
  AND bs.legacy_box_type IN ('DEEP', 'SUPER', 'LARGE_HORIZONTAL_SECTION')
  AND existing.id IS NULL;

-- +goose Down
DROP TABLE IF EXISTS `box_spec_frame_sources`;
