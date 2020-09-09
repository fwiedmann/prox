all: test lint build run
.PHONY: build
build:
	CGO_ENABLED=0 go build -o prox cmd/main.go

.PHONY: test
test:
	go test -coverprofile testcov -race ./...
	rm testcov

.PHONY: run
run:
	sudo ./prox --static-config ./static.yaml --routes-config ./routes.yaml --tls-config ./tls.yaml --loglevel debug

.PHONY: docker
docker:
	docker build -t prox:dev .
	docker-compose up
