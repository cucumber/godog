.PHONY: test

test:
	mysql -u root -e 'DROP DATABASE IF EXISTS `godog_test`'
	mysql -u root -e 'CREATE DATABASE IF NOT EXISTS `godog_test`'
	mysql -u root godog_test < db.sql
	godog users.feature

