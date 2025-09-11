APP=shorty

.PHONY: run build test clean docker

run:
	@go run ./cmd/shorty

build:
	@go build -o $(APP) ./cmd/shorty

test:
	@go test ./...

clean:
	@rm -f $(APP)

docker:
	docker build -t $(APP):latest .
