all: build run
.PHONY: build
build:
	CGO_ENABLED=0 go build -o prox cmd/main.go

.PHONY: run
run:
	./prox --static-config ./static.yaml --routes-config ./routes.yaml