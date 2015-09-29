BRANCH=`git rev-parse --abbrev-ref HEAD`
COMMIT=`git rev-parse --short HEAD`
VERSION=`git describe --always --dirty=-hacky`
GOLDFLAGS="-X main.branch $(BRANCH) -X main.commit $(COMMIT) -X main.version $(VERSION)"
PACKAGENAME=`go list .`

all: test build

setup:
	@echo "==== setup dependencies ==="
	@go get -u "github.com/tools/godep"
	@go get -u "github.com/golang/lint/golint"
	@go get -u "golang.org/x/tools/cmd/vet"
#	@go get -u "github.com/kisielk/errcheck"

# http://cloc.sourceforge.net/
cloc:
	@cloc --sdir='Godeps' --not-match-f='Makefile|_test.go' .

#errcheck:
#	@echo "=== errcheck ==="
#	@errcheck $(PACKAGENAME)/...

vet:
	@echo "==== go vet ==="
	@go vet ./...

lint:
	@echo "==== go lint ==="
	@golint ./**/*.go

fmt:
	@echo "=== go fmt ==="
	@go fmt ./...

install: test
	@echo "=== go install ==="
	@godep go install -ldflags=$(GOLDFLAGS)

build:
	@echo "=== go build ==="
	@godep go build -ldflags=$(GOLDFLAGS)

test: fmt vet lint errcheck
	@echo "=== go test ==="
	@godep go test ./... -cover

deploy: test
	GOARCH=amd64 GOOS=linux godep go build -ldflags=$(GOLDFLAGS)

.PHONY: setup cloc errcheck vet lint fmt install build test
