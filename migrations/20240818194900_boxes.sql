-- +goose Up
ALTER TABLE `boxes` CHANGE `type` `type`
ENUM('SUPER','DEEP','GATE','VENTILATION','QUEEN_EXCLUDER','HORIZONTAL_FEEDER')  CHARACTER SET utf8mb4  COLLATE utf8mb4_0900_ai_ci  NOT NULL  DEFAULT 'DEEP';
