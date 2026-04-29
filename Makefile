GO_BIN := $(shell which go)

PROJECT_NAME := $(shell basename $(CURDIR))

run:
	$(GO_BIN) run ./main.go

build:
	$(GO_BIN) build -o build/$(PROJECT_NAME) ./main.go

clean:
	@rm -rf ./build

update:
	@$(GO_BIN) get -u
	@$(GO_BIN) mod tidy
