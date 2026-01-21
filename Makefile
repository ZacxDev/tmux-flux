.PHONY: build install clean run

BINARY := tmux-flux
GO := go

build:
	$(GO) build -o $(BINARY) ./cmd/tmux-flux/

install: build
	cp $(BINARY) $(HOME)/.local/bin/

clean:
	rm -f $(BINARY)

run: build
	./$(BINARY)

tidy:
	$(GO) mod tidy

test:
	$(GO) test ./...
