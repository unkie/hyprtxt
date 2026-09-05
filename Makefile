APP = hyprtxt
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || printf devel)
GO_LDFLAGS = -X main.version=$(VERSION)
GO_BUILD_FLAGS ?=
PREFIX ?= /usr/local

.PHONY: build install run clean fmt test

build:
	go build $(GO_BUILD_FLAGS) -ldflags "$(GO_LDFLAGS)" -o $(APP) .

install: build
	install -Dm755 "$(APP)" "$(DESTDIR)$(PREFIX)/bin/$(APP)"
	install -Dm644 README.md "$(DESTDIR)$(PREFIX)/share/doc/$(APP)/README.md"
	install -Dm644 LICENSE "$(DESTDIR)$(PREFIX)/share/licenses/$(APP)/LICENSE"
	install -Dm644 completions/hyprtxt.bash \
		"$(DESTDIR)$(PREFIX)/share/bash-completion/completions/hyprtxt"
	install -Dm644 completions/_hyprtxt \
		"$(DESTDIR)$(PREFIX)/share/zsh/site-functions/_hyprtxt"

run:
	go run -ldflags "$(GO_LDFLAGS)" . $(ARGS)

clean:
	rm -f $(APP)

fmt:
	gofmt -w *.go

test:
	go test ./...
