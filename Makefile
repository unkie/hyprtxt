APP = hyprtxt

.PHONY: build install run clean fmt

build:
	go build -o $(APP) .

install:
	go install .

run:
	go run . $(ARGS)

clean:
	rm -f $(APP)

fmt:
	gofmt -w *.go

