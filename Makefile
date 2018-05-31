NAME := weatherline
VERSION := 1.0.1
REPO := github.com/yyotti/weatherline

SRCS := $(shell find . -type f -name '*.go' -not -name '*_test.go')
DEST_DIR := dest

GO_CMD := go
DEP_CMD := dep
GET_CMD := $(GO_CMD) get -v
BUILD_CMD := $(GO_CMD) build -v -a -tags "netgo" -installsuffix "netgo"
CLEAN_CMD := $(GO_CMD) clean
TEST_CMD := $(GO_CMD) test -v

BIN_NAME := $(NAME)
LDFLAGS := -ldflags="-s -w -X \"$(REPO)/$(NAME)/cmd.version=$(VERSION)\" -extldflags \"static\""

all: test build

.PHONY: dep
dep:
ifeq ($(shell command -v dep 2>/dev/null),)
	$(GET_CMD) github.com/golang/dep
endif

.PHONY: deps
deps: dep
	$(DEP_CMD) ensure

.PHONY: test
test:
	$(TEST_CMD) -cover $(REPO)/...

.PHONY: install
install:
	go install $(LDFLAGS)

.PHONY: clean
clean:
	$(CLEAN_CMD)
	rm -rf $(DEST_DIR)

build: deps
	$(BUILD_CMD) $(LDFLAGS) -o $(DEST_DIR)/$(shell go env GOOS)-$(shell go env GOARCH)/$(NAME) $(REPO)/$(NAME)

.PHONY: cross-build
cross-build: deps
	for os in linux windows; do \
		for arch in amd64 386; do \
			GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 $(BUILD_CMD) $(LDFLAGS) -o $(DEST_DIR)/$$os-$$arch/$(NAME) $(REPO)/$(NAME); \
		done; \
	done
