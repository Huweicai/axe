.PHONY: build install test clean

build:
	go build -o axe .

install:
	go install .

test:
	go test ./... -v

clean:
	rm -f axe
