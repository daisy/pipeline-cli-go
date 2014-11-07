BUILDDIR := ${CURDIR}/build
GOPATH   := ${BUILDDIR}
GO       := env GOPATH="${GOPATH}" go
GOX      := env GOPATH="${GOPATH}" ${GOPATH}/bin/gox


define HELP_TEXT
Available targets:
  help                this help
  clean               clean up
  all                 build binaries
  build               build all
  build-dp2           build dp2 tool
  build-dp2admin      build dp2admin tool
  dist                build x-platform binaries
  cover               run test coverage
  cover-deploy        deploy test coverage results
endef
export HELP_TEXT

.PHONY: help clean build-setup build dist test cover-deploy all

all: build

help:
	@echo "$$HELP_TEXT"

clean:
	-rm -rf "${BUILDDIR}"

build: test build-dp2 build-dp2admin

build-setup:
	@export GOPATH="${GOPATH}"
	@echo "Getting dependencies..."
	@mkdir -p "${GOPATH}/src/github.com/daisy"
	@test -d "${GOPATH}/src/github.com/daisy/pipeline-cli-go" || ln -s "${CURDIR}" "${GOPATH}/src/github.com/daisy/pipeline-cli-go"
	@${GO} get github.com/capitancambio/go-subcommand
	@${GO} get launchpad.net/goyaml 
	@${GO} get github.com/daisy/pipeline-clientlib-go
	@${GO} get bitbucket.org/kardianos/osext
	@${GO} get code.google.com/p/go.tools/cmd/cover 

build-dp2: build-setup
	@echo "Building dp2..."
	@${GO} install ${GOBUILD_FLAGS} github.com/daisy/pipeline-cli-go/dp2

build-dp2admin: build-setup
	@echo "Building dp2admin..."
	@${GO} install ${GOBUILD_FLAGS} github.com/daisy/pipeline-cli-go/dp2admin

dist: build-setup cover
	@echo "Building for x-platform..."
	@${GO} get github.com/mitchellh/gox
	@${GOX} -build-toolchain \
	        -osarch="linux/amd64 linux/386 darwin/386 darwin/amd64 windows/386 windows/amd64"
	@${GOX} -output="${GOPATH}/bin/{{.OS}}_{{.Arch}}/{{.Dir}}" \
	        -osarch="linux/amd64 linux/386 darwin/386 darwin/amd64 windows/386 windows/amd64" \
	        ./dp2/ ./dp2admin

test: build-setup
	@echo "Running tests..."
	@${GO} test -covermode=atomic -coverprofile=${BUILDDIR}/profile.cov ./cli/...

cover-deploy: cover
	@${GO} get github.com/modocache/gover
	@${GO} get github.com/mattn/goveralls
	GOPATH="${GOPATH}" ${GOPATH}/bin/goveralls \
	      -coverprofile=${BUILDDIR}/profile.cov \
	      -service=travis-ci
