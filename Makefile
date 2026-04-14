.PHONY: build clean run

BINARY_NAME=fuzzer

build:
	@echo "Building fuzzer..."
	go build -o bin/$(BINARY_NAME) cmd/fuzzer/main.go

run: build
	./bin/$(BINARY_NAME) --target 127.0.0.1 --ports 1024

clean:
	go clean
	rm -rf bin/
	rm -f results.jsonl
