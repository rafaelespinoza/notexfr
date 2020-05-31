GO ?= go
PKG_IMPORT_PATH=github.com/rafaelespinoza/notexfr
BIN_DIR=./bin

BRANCH_NAME=$(shell git rev-parse --abbrev-ref HEAD)
BUILD_TIME=$(shell date --rfc-3339=seconds --utc | tr ' ' 'T')
COMMIT_HASH=$(shell git rev-parse --short=7 HEAD)
GO_OS_ARCH=$(shell $(GO) version | awk '{ print $$4 }' | tr '/' '_')
GO_VERSION=$(shell $(GO) version | awk '{ print $$3 }')
RELEASE_TAG=$(shell git describe --tags)

build:
	mkdir -pv $(BIN_DIR)
	$(GO) build \
		-o $(BIN_DIR)/notexfr-$(RELEASE_TAG)-$(GO_OS_ARCH) \
		-v -ldflags "\
			-X $(PKG_IMPORT_PATH)/internal/version.BranchName=$(BRANCH_NAME) \
			-X $(PKG_IMPORT_PATH)/internal/version.BuildTime=$(BUILD_TIME) \
			-X $(PKG_IMPORT_PATH)/internal/version.CommitHash=$(COMMIT_HASH) \
			-X $(PKG_IMPORT_PATH)/internal/version.GoOSArch=$(GO_OS_ARCH) \
			-X $(PKG_IMPORT_PATH)/internal/version.GoVersion=$(GO_VERSION) \
			-X $(PKG_IMPORT_PATH)/internal/version.ReleaseTag=$(RELEASE_TAG)"

clean:
	rm -frv $(BIN_DIR)
