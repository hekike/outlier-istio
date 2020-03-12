#
# Tools and binaries
#
GOCMD		= go
GOTEST		=$(GOCMD) test
PROTOC		= protoc

#
# Directories and packages
#
TEST_PKGS := $(shell go list ./... | grep -v fixtures)
GOOGLEAPIS_DIR := ../googleapis

#
# Targets
#
.PHONY: test
test:
	$(GOTEST) $(TEST_PKGS)
.PHONY: testv
testv:
	$(GOTEST) -v $(TEST_PKGS)
