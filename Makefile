

default: test

run:
	go run go-house.go --config_dir $(SANDBOX_DIR)/config

test:
	go test -timeout 10s ./...

network:
	cd spark-api && go test -network .

lint:
	gofmt -s -l .
	go vet ./...

install:
	go install .
	sudo install -d /usr/local/go-house/ /usr/local/go-house/static/
	sudo install -t /usr/local/go-house/ $(GOPATH)/bin/go-house
	sudo install -t /usr/local/go-house/static/ $(GOPATH)/static/*
	sudo /etc/init.d/go-house restart

clean:
	go clean
