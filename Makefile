GO ?= go

PKG_IMPORT_PATH=github.com/rafaelespinoza/notexfr
BIN_DIR=./bin
MAIN=$(BIN_DIR)/main
SRC_PATHS = . ./internal/...

BRANCH_NAME=$(shell git rev-parse --abbrev-ref HEAD)
BUILD_TIME=$(shell date --rfc-3339=seconds --utc | tr ' ' 'T')
COMMIT_HASH=$(shell git rev-parse --short=7 HEAD)
GO_OS_ARCH=$(shell $(GO) version | awk '{ print $$4 }' | tr '/' '_')
GO_VERSION=$(shell $(GO) version | awk '{ print $$3 }')
RELEASE_TAG=$(shell git describe --tags)

build:
	mkdir -pv $(BIN_DIR)
	$(GO) build \
		-o $(BIN_DIR)/notexfr \
		-v -ldflags "\
			-X $(PKG_IMPORT_PATH)/internal/version.BranchName=$(BRANCH_NAME) \
			-X $(PKG_IMPORT_PATH)/internal/version.BuildTime=$(BUILD_TIME) \
			-X $(PKG_IMPORT_PATH)/internal/version.CommitHash=$(COMMIT_HASH) \
			-X $(PKG_IMPORT_PATH)/internal/version.GoOSArch=$(GO_OS_ARCH) \
			-X $(PKG_IMPORT_PATH)/internal/version.GoVersion=$(GO_VERSION) \
			-X $(PKG_IMPORT_PATH)/internal/version.ReleaseTag=$(RELEASE_TAG)"

deps:
	$(GO) mod tidy && $(GO) mod vendor

# Specify packages to test with P variable. Example:
# make test P='entity repo'
#
# Specify test flags with FLAGS variable. Example:
# make test FLAGS='-v -count=1 -failfast'
test: P ?= ...
test: pkgpath=$(foreach pkg,$(P),$(shell echo ./internal/$(pkg)))
test:
test:
	$(GO) test $(pkgpath) $(FLAGS)

vet:
	$(GO) vet $(FLAGS) $(SRC_PATHS)
