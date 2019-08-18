-- +migrate Up
CREATE TABLE `user_account` (
  `id` VARCHAR(50) PRIMARY KEY,
  `email` VARCHAR(64) NULL,
  `created_at` DATE NULL
);
-- +migrate Down
DROP TABLE IF EXISTS `user_account`;