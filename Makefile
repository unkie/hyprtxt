APP = hyprtxt

build:
	go build -o $(APP) main.go font.go

install:
	go install ./...

run:
	go run main.go font.go -- "$(ARGS)"

clean:
	rm -f $(APP)
