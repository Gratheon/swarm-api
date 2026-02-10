-- +goose Up
/* 20:01:23 local swarm-api */ ALTER TABLE `frames_sides` DROP `brood`;
/* 20:02:03 local swarm-api */ ALTER TABLE `frames_sides` DROP `capped_brood`;
/* 20:02:18 local swarm-api */ ALTER TABLE `frames_sides` DROP `eggs`;
/* 20:02:31 local swarm-api */ ALTER TABLE `frames_sides` DROP `pollen`;
/* 20:02:44 local swarm-api */ ALTER TABLE `frames_sides` DROP `honey`;

/* 20:13:33 local swarm-api */ ALTER TABLE `frames_sides` DROP `queen_detected`;
