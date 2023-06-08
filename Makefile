GOCMD=go
GOBUILD=$(GOCMD) build
BINARY_NAME=./bin/hydro

all: test build run_server

wire_gen:
	wire gen ./cmd/hydro
	echo "wire build"

build:
	$(GOBUILD) -o $(BINARY_NAME) ./cmd/hydro
	echo "binary build"

run_server:
	LOG_LEVEL=debug $(BINARY_NAME)

docker_all: wire_build docker_build

docker_build:
	docker build -t kara/hydro:latest

test:
	$(GOCMD) test -v ./...