#
# Tools and binaries
#
GOCMD		= go
GOTEST		=$(GOCMD) test

#
# Directories and packages
#
TEST_PKGS := $(shell go list ./... | grep /test | grep -v fixtures)

#
# Targets
#
.PHONY: test
test:
	$(GOTEST) $(TEST_PKGS)
.PHONY: testv
testv:
	$(GOTEST) -v $(TEST_PKGS)
