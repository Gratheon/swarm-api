-- +goose Up
/* 17:19:48 local swarm-api */
ALTER TABLE `hives` ADD `added` DATETIME NULL AFTER `notes`;
/* 17:22:44 local swarm-api */
ALTER TABLE `hives`
ADD `status` VARCHAR(50) NULL DEFAULT 'active' AFTER `added`;

CREATE TABLE `treatments` (
    `id` int unsigned NOT NULL AUTO_INCREMENT,
    `user_id` int DEFAULT NULL,
    `box_id` int DEFAULT NULL,
    `hive_id` int DEFAULT NULL,
    `family_id` int NOT NULL,
    `added` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `type` varchar(50) COLLATE utf8mb4_general_ci DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci;