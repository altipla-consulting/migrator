
FILES = $(shell find . -type f -name '*.go')

gofmt:
	@gofmt -w $(FILES)
	@gofmt -r '&α{} -> new(α)' -w $(FILES)

deps:
	go get -u github.com/mgechev/revive

test: gofmt
	revive -formatter friendly
	go install ./cmd/init-migrator
	go install ./cmd/migrator

	docker-compose kill database
	docker-compose rm -f database
	docker-compose up -d database
	bash -c "until mysql -h 127.0.0.1 -P 3307 -u dev-user -pdev-password -e ';' 2> /dev/null ; do sleep 1; done"

	init-migrator -user root -password dev-root -address localhost:3307
	migrator -user root -password dev-root -address localhost:3307 -directory testdata/full
	migrator -user root -password dev-root -address localhost:3307 -directory testdata/partial -start 002_example_migration.sql

update-deps:
	go get -u
	go mod download
	go mod tidy
