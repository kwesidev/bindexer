all: clean build

run:
	go run main.go

build:
	go build -ldflags="-s -w" -trimpath

clean:
	rm -rf bindexer
