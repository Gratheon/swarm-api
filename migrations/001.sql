
SET FOREIGN_KEY_CHECKS=0;

CREATE TABLE
    IF NOT EXISTS 'apiaries' (
        'id' int unsigned NOT NULL AUTO_INCREMENT,
        'user_id' int unsigned NOT NULL,
        'name' varchar(250) DEFAULT NULL,
        'active' tinyint(1) NOT NULL DEFAULT 1,
        'lng' varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT "0",
        'lat' varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT "0",
        PRIMARY KEY ('id')
    ) ENGINE = InnoDB AUTO_INCREMENT = 2 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci;

CREATE TABLE
    IF NOT EXISTS 'hives' (
        'id' int unsigned NOT NULL AUTO_INCREMENT,
        'user_id' int unsigned NOT NULL,
        'family_id' int unsigned DEFAULT NULL,
        'apiary_id' int unsigned DEFAULT NULL,
        'name' varchar(250) DEFAULT NULL,
        'notes' mediumtext DEFAULT NULL,
        'color' varchar(20) DEFAULT NULL,
        'active' tinyint(1) NOT NULL DEFAULT 1,
        PRIMARY KEY ('id')
    ) ENGINE = InnoDB AUTO_INCREMENT = 1 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci;

CREATE TABLE
    IF NOT EXISTS 'boxes' (
        'id' int unsigned NOT NULL AUTO_INCREMENT,
        'user_id' int unsigned NOT NULL,
        'hive_id' int NOT NULL,
        'active' tinyint(1) NOT NULL DEFAULT 1,
        'color' varchar(10) DEFAULT NULL,
        'position' mediumint DEFAULT NULL,
        'type' enum("SUPER", "DEEP") COLLATE utf8mb4_general_ci NOT NULL DEFAULT "DEEP",
        PRIMARY KEY ('id')
    ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci;

CREATE TABLE
    IF NOT EXISTS 'families' (
        'id' int unsigned NOT NULL AUTO_INCREMENT,
        'user_id' int DEFAULT NULL,
        'race' varchar(100) DEFAULT NULL,
        'added' varchar(4) DEFAULT NULL,
        PRIMARY KEY ('id')
    ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci;

CREATE TABLE
    IF NOT EXISTS 'frames_sides' (
        'id' int unsigned NOT NULL AUTO_INCREMENT,
        'user_id' int unsigned NOT NULL,
        'brood' int DEFAULT NULL,
        'capped_brood' int DEFAULT NULL,
        'eggs' int DEFAULT NULL,
        'pollen' int DEFAULT NULL,
        'honey' int DEFAULT NULL,
        'queen_detected' tinyint(1) NOT NULL DEFAULT 0,
        PRIMARY KEY ('id')
    ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci;

CREATE TABLE
    IF NOT EXISTS 'frames' (
        'id' int unsigned NOT NULL AUTO_INCREMENT,
        'user_id' int unsigned NOT NULL,
        'box_id' int unsigned DEFAULT NULL,
        'type' enum(
            "VOID",
            "FOUNDATION",
            "EMPTY_COMB",
            "PARTITION",
            "FEEDER"
        ) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT "EMPTY_COMB",
        'position' int unsigned DEFAULT NULL,
        'left_id' int unsigned DEFAULT NULL,
        'right_id' int unsigned DEFAULT NULL,
        'active' tinyint(1) NOT NULL DEFAULT 1,
        PRIMARY KEY ('id'),
        KEY 'box_id' ('box_id'),
        KEY 'left_id' ('left_id'),
        KEY 'right_id' ('right_id'),
        CONSTRAINT 'frames_ibfk_1' FOREIGN KEY ('box_id') REFERENCES 'boxes' ('id') ON DELETE CASCADE,
        CONSTRAINT 'frames_ibfk_2' FOREIGN KEY ('left_id') REFERENCES 'frames_sides' ('id') ON DELETE
        SET
            NULL,
            CONSTRAINT 'frames_ibfk_3' FOREIGN KEY ('right_id') REFERENCES 'frames_sides' ('id') ON DELETE
        SET
            NULL
    ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci;

CREATE TABLE
    IF NOT EXISTS 'inspections' (
        'id' int unsigned NOT NULL AUTO_INCREMENT,
        'hive_id' int DEFAULT NULL,
        'user_id' int unsigned NOT NULL,
        'data' JSON,
        'added' datetime DEFAULT NULL,
        PRIMARY KEY ('id')
    ) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci;

SET FOREIGN_KEY_CHECKS=1;