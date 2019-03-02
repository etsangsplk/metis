gofiles     = $(shell find . -type f -iname '*.go' | grep -v vendor)
now         = $(shell date -u)
git_rev     = $(shell git rev-parse HEAD)
version     = $(shell git describe --tags --abbrev=0 2>/dev/null || echo -n $(git_rev))
GO_VERSION := $(shell cat .go-version)

mkdir       = @mkdir -p $(dir $@)
print       = @printf ":::::::::::::::: [$(now)] $@ ::::::::::::::::\n"
touch       = @touch $@

GOOS       ?= linux
GOARCH     ?= amd64

ldflags = '\
	-X "main.version=$(version)" \
	-X "main.revision=$(git_rev)" \
	-X "main.buildDate=$(now)" \
	-extldflags "-static" \
'

go = docker run --rm \
	-u $(shell id -u) \
	-v "$(CURDIR):$(CURDIR)" \
	-e "GOOS=$(GOOS)" \
	-e "GOARCH=$(GOARCH)" \
	-e "CGO_ENABLED=0" \
	-e "GOCACHE=$(CURDIR)/target/.cache/go" \
	-e "GO111MODULE=on" \
	-e "GOFLAGS=-mod=vendor" \
	-w "$(CURDIR)" \
	golang:$(GO_VERSION) \
	go

linter = @docker run \
	--rm -i \
	-e "GO111MODULE=on" \
	-e "GOFLAGS=-mod=vendor" \
	-e "GOCACHE=$(CURDIR)/target/.cache/go" \
	-v "$(CURDIR):$(CURDIR)" \
	-w "$(CURDIR)" \
	-u $(shell id -u) \
	golangci/golangci-lint:v1.15.0 \
	golangci-lint

target/metis-$(GOOS)-$(GOARCH): $(gofiles)
	$(print)
	$(mkdir)
	@$(go) build -a \
		-installsuffix cgo \
		-ldflags $(ldflags) \
		-o $@ \
		./cmd/metis-store/*.go

docker: target/.cache/docker-image-$(git_rev)
target/.cache/docker-image-$(git_rev): Dockerfile $(gofiles)
	$(mkdir)
	@docker build . \
		-t docker.io/digitalocean/metis:$(version)
	$(touch)

lint: target/.cache/linter
target/.cache/linter: $(gofiles)
	$(mkdir)
	$(print)
	$(linter) run
	$(touch)

test: target/.coverprofile
target/.coverprofile:
	$(mkdir)
	$(print)
	@$(go) test -coverprofile=$@ ./...

.PHONY: ci
ci: lint docker test