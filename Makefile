APP = hyprtxt
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || printf devel)
GO_LDFLAGS = -X main.version=$(VERSION)

.PHONY: build install run clean fmt

build:
	go build -ldflags "$(GO_LDFLAGS)" -o $(APP) .

install:
	go install -ldflags "$(GO_LDFLAGS)" .

run:
	go run -ldflags "$(GO_LDFLAGS)" . $(ARGS)

clean:
	rm -f $(APP)

fmt:
	gofmt -w *.go

