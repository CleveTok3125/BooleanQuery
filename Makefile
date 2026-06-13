PREFIX  ?= /usr/local
BINDIR  ?= $(PREFIX)/bin
BINARY  := bq

all: build

build:
	go build -o $(BINARY) ./src/

install: build
	install -d $(DESTDIR)$(BINDIR)
	install -m 755 $(BINARY) $(DESTDIR)$(BINDIR)/$(BINARY)

uninstall:
	rm -f $(DESTDIR)$(BINDIR)/$(BINARY)

clean:
	rm -f $(BINARY)

test:
	go test ./...

fmt:
	go fmt ./...

fmtcheck:
	test -z "$$(gofmt -l .)"

vet:
	go vet ./...

tidy:
	go mod tidy
	git diff --exit-code go.mod go.sum

check: fmtcheck vet build tidy

.PHONY: all build install uninstall clean test fmt fmtcheck vet tidy check
