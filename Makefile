

default: test

run:
	go run go-house.go --config_dir $(SANDBOX_DIR)/config --static_dir $(SANDBOX_DIR)/static

test:
	go test ./...

clean:
	go clean
