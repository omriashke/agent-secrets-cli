.PHONY: build install clean lint release snapshot

BIN := agent-secrets
MAIN := ./main.go

build:
	go build -o $(BIN) $(MAIN)

install: build
	mv $(BIN) /usr/local/bin/$(BIN)

clean:
	rm -f $(BIN)
	rm -rf dist/

lint:
	go vet ./...

snapshot:
	goreleaser release --snapshot --clean

release:
	goreleaser release --clean
