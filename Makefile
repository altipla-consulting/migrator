
FILES = $(shell find . -type f -name "*.go" -not -path "./vendor/*")

gofmt:
	@gofmt -w $(FILES)
	@gofmt -r '&a{} -> new(a)' -w $(FILES)

deps:
	go get -u github.com/mgechev/revive

test:
	@./infra/test.sh

test-travis:
	@./infra/test-travis.sh
