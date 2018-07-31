REVISION := $(shell git describe --always)
DATE := $(shell date +%Y-%m-%dT%H:%M:%S%z)
LDFLAGS := -ldflags="-X \"main.Revision=$(REVISION)\" -X \"main.BuildDate=${DATE}\""

.PHONY: build-cross dist build dep dep/update clean run help

name            := deliverbot
linux_name      := $(name)-linux-amd64
darwin_name     := $(name)-darwin-amd64

build-cross: ## create to build for linux & darwin to bin/
	GOOS=linux GOARCH=amd64 go build -o bin/$(linux_name) $(LDFLAGS) *.go
	GOOS=darwin GOARCH=amd64 go build -o bin/$(darwin_name) $(LDFLAGS) *.go

dist: build-cross ## create .tar.gz linux & darwin to /bin
	cd bin && tar zcvf $(linux_name).tar.gz $(linux_name) && rm -f $(linux_name)
	cd bin && tar zcvf $(darwin_name).tar.gz $(darwin_name) && rm -f $(darwin_name)

build: ## go build
	go build -o bin/$(name) $(LDFLAGS) *.go

test: ## go test
	go test -v $$(go list ./... | grep -v /vendor/)

dep/install:
ifeq ($(which shell dep),)
	go get -u github.com/golang/dep/cmd/dep
endif

dep: dep/install ## dep ensure
	dep ensure

dep/init: dep/install ## dep init
	dep init

dep/update: dep/install ## dep update
	dep ensure -update

clean: ## remove bin/*
	rm -f bin/*

run: ## go run
	go run main.go -c examples/config.toml

help:
	@awk -F ':|##' '/^[^\t].+?:.*?##/ { printf "\033[36m%-22s\033[0m %s\n", $$1, $$NF }' $(MAKEFILE_LIST)
