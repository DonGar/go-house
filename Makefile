

default: test

run:
	go run go-house.go --config_dir $(SANDBOX_DIR)/config

test:
	go test ./...

clean:
	go clean
