RELEASE_VERSION    = $(release_version)

ifeq ("$(RELEASE_VERSION)","")
	RELEASE_VERSION		:= "unknown"
endif

ROOT_DIR 		   = $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))/
VERSION_PATH	   = $(shell echo $(ROOT_DIR) | sed -e "s;${GOPATH}/src/;;g")pkg/util
LD_GIT_COMMIT      = -X '$(VERSION_PATH).GitCommit=`git rev-parse --short HEAD`'
LD_BUILD_TIME      = -X '$(VERSION_PATH).BuildTime=`date +%FT%T%z`'
LD_GO_VERSION      = -X '$(VERSION_PATH).GoVersion=`go version`'
LD_VERSION = -X '$(VERSION_PATH).Version=$(RELEASE_VERSION)'
LD_FLAGS           = -ldflags "$(LD_GIT_COMMIT) $(LD_BUILD_TIME) $(LD_GO_VERSION) $(LD_VERSION) -w -s"

GOOS 		= linux
CGO_ENABLED = 0
DIST_DIR 	= $(ROOT_DIR)dist/


.PHONY: release
release: dist_dir converthouse;

.PHONY: release_darwin
release_darwin: darwin dist_dir converthouse;

.PHONY: darwin
darwin:
	$(eval GOOS := darwin)

.PHONY: docker
docker:
	@echo ========== current docker tag is: $(RELEASE_VERSION) ==========
	docker build --build-arg RELEASE=$(RELEASE_VERSION) -t deepfabric/converthouse:$(RELEASE_VERSION) -f Dockerfile .
	docker tag deepfabric/converthouse:$(RELEASE_VERSION) deepfabric/converthouse

.PHONY: converthouse
converthouse: ; $(info ======== compiled converthouse:)
	env GO111MODULE=on CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) go build -mod=vendor -a -installsuffix cgo -o $(DIST_DIR)converthouse $(LD_FLAGS) $(ROOT_DIR)cmd/converthouse/*.go

.PHONY: dist_dir
dist_dir: ; $(info ======== prepare distribute dir:)
	mkdir -p $(DIST_DIR)
	@rm -rf $(DIST_DIR)*

.PHONY: clean
clean: ; $(info ======== clean all:)
	rm -rf $(DIST_DIR)*

.PHONY: help
help:
	@echo "build release binary: \n\t\tmake release\n"
	@echo "build Mac OS X release binary: \n\t\tmake release_darwin\n"
	@echo "build docker release: \n\t\tmake docker\n"
	@echo "\t  default: all, like 「make docker」\n"
	@echo "\t  converthouse: compile converthouse\n"
	@echo "clean all binary: \n\t\tmake clean\n"

UNAME_S := $(shell uname -s)

ifeq ($(UNAME_S),Darwin)
	.DEFAULT_GOAL := release_darwin
else
	.DEFAULT_GOAL := release
endif
