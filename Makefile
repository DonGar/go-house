

default: test

run:
	go run go-house.go

test:
	go test ./...

clean:
	go clean
