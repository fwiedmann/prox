all: test lint build run
.PHONY: build
build:
	CGO_ENABLED=0 go build -o prox cmd/main.go

.PHONY: test
test:
	go test -coverprofile testcov -race ./...
	rm testcov

.PHONY: lint
lint:
	golangci-lint run

.PHONY: run
run:
	sudo ./prox --static-config ./static.yaml --routes-config ./routes.yaml --tls-config ./tls.yaml

.PHONY: docker
docker:
	docker build -t prox:dev .
