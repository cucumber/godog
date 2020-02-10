
define DB_SQL
CREATE TABLE users (
  `id` BIGINT UNSIGNED AUTO_INCREMENT NOT NULL,
  `username` VARCHAR(32) NOT NULL,
  `email` VARCHAR(255) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uniq_email` (`email`)
) DEFAULT CHARACTER SET utf8 COLLATE utf8_unicode_ci ENGINE = InnoDB;
endef

export DB_SQL

SQL := "$$DB_SQL"

test:
	mysql -u root -e 'DROP DATABASE IF EXISTS `godog_test`'
	mysql -u root -e 'CREATE DATABASE IF NOT EXISTS `godog_test`'
	@mysql -u root godog_test -e $(SQL)
	godog users.feature

.PHONY: test
