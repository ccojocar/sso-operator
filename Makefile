SHELL := /bin/bash
GO := GO15VENDOREXPERIMENT=1 go
NAME := sso-operator
OS := $(shell uname)
MAIN_GO := main.go
ROOT_PACKAGE := $(GIT_PROVIDER)/$(ORG)/$(NAME)
GO_VERSION := $(shell $(GO) version | sed -e 's/^[^0-9.]*\([0-9.]*\).*/\1/')
PACKAGE_DIRS := $(shell $(GO) list ./... | grep -v /vendor/)
PKGS := $(shell go list ./... | grep -v /vendor | grep -v generated)
PKGS := $(subst  :,_,$(PKGS))
BUILDFLAGS := ''
CGO_ENABLED = 0
VENDOR_DIR=vendor

all: bootstrap fmt lint sec test build

DEP := $(GOPATH)/bin/dep
$(DEP):
	go get -u github.com/golang/dep/cmd/dep

bootstrap: $(DEP)
	@echo "INSTALLING DEPENDENCIES"
	$(DEP) ensure 

build:
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build -ldflags $(BUILDFLAGS) -o bin/$(NAME) $(MAIN_GO)

test: 
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test $(PACKAGE_DIRS) -test.v

install:
	GOBIN=${GOPATH}/bin $(GO) install -ldflags $(BUILDFLAGS) $(MAIN_GO)

fmt:
	@echo "FORMATTING"
	@FORMATTED=`$(GO) fmt $(PACKAGE_DIRS)`
	@([[ ! -z "$(FORMATTED)" ]] && printf "Fixed unformatted files:\n$(FORMATTED)") || true

clean:
	rm -rf build release $(VENDOR_DIR)

GOLINT := $(GOPATH)/bin/golint
$(GOLINT):
	go get -u github.com/golang/lint/golint

.PHONY: lint
lint: $(GOLINT)
	@echo "VETTING"
	go vet $(go list ./... | grep -v /vendor/)
	@echo "LINTING"
	$(GOLINT) -set_exit_status $(shell go list ./... | grep -v vendor)

GOSEC := $(GOPATH)/bin/gosec
$(GOSEC):
	go get -u github.com/securego/gosec/cmd/gosec/...

.PHONY: sec
sec: $(GOSEC)
	@echo "SECURITY SCANNING"
	$(GOSEC) -fmt=csv ./...

watch:
	reflex -r "\.go$" -R "vendor.*" make skaffold-run

skaffold-run: build
	skaffold run -p dev
