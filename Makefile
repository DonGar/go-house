

default: test

run:
	go run go-house.go --config_dir $(SANDBOX_DIR)/config

test:
	go test ./...

lint:
	gofmt -s -l .
	go vet ./...

install:
	go install .
	sudo install -d /usr/local/go-house/ /usr/local/go-house/static/
	sudo install -t /usr/local/go-house/ $(GOPATH)/bin/go-house
	sudo install -t /usr/local/go-house/static/ $(GOPATH)/static/*

clean:
	go clean
