REPO := github.com/Code-Hex/shibafu-playground
GIT_REF := $(shell git describe --always --tag)
VERSION ?= $(GIT_REF)

.PHONY: build
build:
	@echo "+ $@"
	CGO_ENABLED=0 go build -o bin/server \
        -ldflags "-w -s -X main.version=$(VERSION)" \
        $(REPO)

.PHONY: build/alpine
build/alpine:
	@echo "+ $@"
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/server \
        -ldflags "-w -s -X main.version=$(VERSION)" \
        $(REPO)


