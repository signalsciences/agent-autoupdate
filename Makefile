GIT_COMMIT := $(shell git log --pretty=format:'%H' -n 1)
VERSION := $(shell cat ./VERSION)

.DEFAULT_GOAL := build

.PHONY: fmt build msi lint syso

lint:
	GOOS=windows GOARCH=amd64 golangci-lint run

fmt:
	go fmt ./...

build: lint syso
	GOOS=windows GOARCH=amd64 go build -ldflags "-X 'main.buildSha=$(GIT_COMMIT)'" -o ./packaging/agent-autoupdate.exe

msi: build fmt
	export VERSION=$(VERSION); \
	wixl -a x64 ./packaging/installer.wxs -o ./packaging/agent-autoupdate-$(VERSION).msi


syso: # creates a syso file which contains Microsoft Version Information.
	export VERSION=$(VERSION); \
	go run tools/msverinfo/main.go \
		-icon="./packaging/sigsci.ico" \
		-version=$(VERSION) \
		-buildsha=$(shell git rev-parse --short HEAD)