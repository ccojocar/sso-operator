SHELL := /bin/bash
NAME := sso-operator
OS := $(shell uname)
MAIN_GO := main.go
GO := GO111MODULE=on go
GO_NOMOD :=GO111MODULE=off go
BUILDFLAGS := ''
CGO_ENABLED = 0
GOPATH ?= $(shell $(GO) env GOPATH)
GOBIN ?= $(GOPATH)/bin
GOLINT ?= $(GOBIN)/golint
GOSEC ?= $(GOBIN)/gosec

all: fmt lint sec test build

build:
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build -ldflags $(BUILDFLAGS) -o bin/$(NAME) $(MAIN_GO)

test: 
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test -test.v ./...

install:
	GOBIN=${GOPATH}/bin $(GO) install -ldflags $(BUILDFLAGS) $(MAIN_GO)

fmt:
	@echo "FORMATTING"
	@FORMATTED=`$(GO) fmt ./...`
	@([[ ! -z "$(FORMATTED)" ]] && printf "Fixed unformatted files:\n$(FORMATTED)") || true

clean:
	rm -rf build release

lint:
	@echo "LINTING"
	$(GO_NOMOD) get -u golang.org/x/lint/golint
	$(GOLINT) -set_exit_status ./...
	@echo "VETTING"
	$(GO) vet ./...

sec:
	@echo "SECURITY SCANNING"
	$(GO_NOMOD) get github.com/securego/gosec/cmd/gosec
	$(GOSEC) ./...

test-coverage:
	go test -race -coverprofile=coverage.txt -covermode=atomic

codegen:
	@echo "GENERATING KUBERNETES CRDs"
	hack/update-codegen.sh

linux:
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 $(GO) build -ldflags $(BUILDFLAGS) -o bin/$(NAME) $(MAIN_GO)

install-helm: linux
	skaffold run -p install -n $(KUBERNETES_NAMESPACE)
