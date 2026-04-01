build:
	CGO_ENABLED=0 go build -o barrage ./cmd/barrage/

run: build
	./barrage

test:
	go test ./...

clean:
	rm -f barrage

.PHONY: build run test clean
