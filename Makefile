default: build

build:
	go build -o build/go-microservice ./internal/app

run:
	go run ./internal/app

test:
	go test ./internal/app/... -v
