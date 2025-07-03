BINARY_NAME=huemidi

.PHONY: build clean run install

build:
	go build -o $(BINARY_NAME)

clean:
	go clean
	rm -f $(BINARY_NAME)

run: build
	./$(BINARY_NAME)

install: build
	mv $(BINARY_NAME) /usr/local/bin/

deps:
	go mod tidy
	go mod download

.DEFAULT_GOAL := build
