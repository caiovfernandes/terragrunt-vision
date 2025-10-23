all: build

build:
	go build -o terragrunt-vision

test:
	go test ./... -v

test-coverage:
	go test ./... -cover

clean:
	rm -f terragrunt-vision
	rm -f main

install:
	go install

run:
	./terragrunt-vision

.PHONY: all build test test-coverage clean install run
