.PHONY: dev build run tidy migrate

dev:
	air

build:
	go build -o ./tmp/main .

run:
	go run .

tidy:
	go mod tidy

migrate:
	go run . --migrate
