migrate:
	atlas schema apply -u "mysql://root:root@grit-mysql:3306/grit" --to "file://cmd/sql/" --dev-url "mysql://root:root@grit-mysql:3306/dev?foreign_key_checks=0"

