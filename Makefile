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
.PHONY: protoc
protoc:
	$(PROTOC) \
		--go_out=plugins=grpc:. \
		--include_imports \
		--include_source_info \
		--proto_path=${GOOGLEAPIS_DIR} \
		--proto_path=. \
        	--descriptor_set_out=./pkg/api/protobuf/api_descriptor.pb \
		./pkg/api/protobuf/healthcheck.proto

.PHONY: test
test:
	$(GOTEST) $(TEST_PKGS)
.PHONY: testv
testv:
	$(GOTEST) -v $(TEST_PKGS)
