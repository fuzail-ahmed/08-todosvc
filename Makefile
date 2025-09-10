.PHONY: all build run migrate proto test clean

BINARY=bin/todosvc
PROTO_DIR=proto

all: build

# build the binary
build:
	@echo "==> building..."
	go build -o $(BINARY) ./cmd/server

# run local (assumes DB running)
run:
	@echo "==> running..."
	./$(BINARY)

# run migrations (the server will auto-migrate on startup too)
migrate:
	@echo "==> running migrations (server AutoMigrate)..."
	./$(BINARY) --migrate-only

# regenerate protos (needs protoc installed)
proto:
	protoc -I=$(PROTO_DIR) \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/todo.proto

test:
	go test ./...

clean:
	@rm -rf $(BINARY)
