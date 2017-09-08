.PHONY: all
all: setup lint

sources = $(shell find . -name '*.go' -not -path './vendor/*')
.PHONY: goimports
goimports: setup
	goimports -w $(sources)

.PHONY: lint
lint: setup
	gometalinter ./... --enable=goimports --disable=gotype --disable=golint --disable=errcheck --vendor -t

.PHONY: errcheck
errcheck: setup
	gometalinter ./... --disable-all --enable=errcheck --vendor -t

.PHONY: install
install: setup
	go install

BIN_DIR := $(GOPATH)/bin
GOIMPORTS := $(BIN_DIR)/goimports
GOMETALINTER := $(BIN_DIR)/gometalinter
DEP := $(BIN_DIR)/dep
GOCOV := $(BIN_DIR)/gocov
GOCOV_HTML := $(BIN_DIR)/gocov-html

$(GOIMPORTS):
	go get -u golang.org/x/tools/cmd/goimports

$(GOMETALINTER):
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install &> /dev/null

$(DEP):
	go get -u github.com/golang/dep/cmd/dep

tools: $(GOIMPORTS) $(GOMETALINTER) $(DEP)

vendor: $(DEP)
	dep ensure

setup: tools vendor

updatedeps:
	dep ensure -update

BINARY := jaal
VERSION ?= latest

.PHONY: linux
linux: setup
	mkdir -p $(CURDIR)/release
	GOOS=linux GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION)" \
	-o release/$(BINARY)-$(VERSION)-linux-amd64

.PHONY: release
release: linux
