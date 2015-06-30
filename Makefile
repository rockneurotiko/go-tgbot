# Makefile for a go project
#
# Author: Jon Eisen
# 	site: joneisen.me
#
# Targets:
# 	all: Builds the code
# 	build: Builds the code
# 	fmt: Formats the source files
# 	clean: cleans the code
# 	install: Installs the code to the GOPATH
# 	iref: Installs referenced projects
#	test: Runs the tests
#
#  Blog post on it: http://joneisen.me/post/25503842796
#

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOINSTALL=$(GOCMD) install
GOTEST=$(GOCMD) test
GOLINT=golint
GODEP=$(GOTEST) -i
GOFMT=gofmt -w
BINARYNAME=gotgbot

# Package lists
TOPLEVEL_PKG := .
# INT_LIST := tgtypes	#<-- Interface directories
# IMPL_LIST := tgtypes	#<-- Implementation directories
CMD_LIST :=  example/manualexample example/echoexample example/simpleexample	#<-- Command directories

# List building
ALL_LIST = $(TOPLEVEL_PKG)  $(CMD_LIST) # $(INT_LIST) $(IMPL_LIST)

BUILD_LIST = $(foreach int, $(ALL_LIST), $(int)_build)
CLEAN_LIST = $(foreach int, $(ALL_LIST), $(int)_clean)
INSTALL_LIST = $(foreach int, $(ALL_LIST), $(int)_install)
IREF_LIST = $(foreach int, $(ALL_LIST), $(int)_iref)
TEST_LIST = $(foreach int, $(ALL_LIST), $(int)_test)
LINT_LIST = $(foreach int, $(ALL_LIST), $(int)_lint)
FMT_TEST = $(foreach int, $(ALL_LIST), $(int)_fmt)

# All are .PHONY for now because dependencyness is hard
.PHONY: $(CLEAN_LIST) $(TEST_LIST) $(FMT_LIST) $(LINT_LIST) $(INSTALL_LIST) $(BUILD_LIST) $(IREF_LIST)

all: build
build: 	$(BUILD_LIST)
clean: 	$(CLEAN_LIST)
	rm $(BINARYNAME)
install: $(INSTALL_LIST)
test: 	$(TEST_LIST)
lint: 	$(LINT_LIST)
iref: 	$(IREF_LIST)
fmt: 	$(FMT_TEST)
run: 	build
	./$(BINARYNAME)

$(BUILD_LIST): %_build: %_fmt %_lint %_iref %_test
	$(GOBUILD) -o $(TOPLEVEL_PKG)/$(BINARYNAME) $(TOPLEVEL_PKG)/$*
$(CLEAN_LIST): %_clean:
	$(GOCLEAN) $(TOPLEVEL_PKG)/$*
$(INSTALL_LIST): %_install:
	$(GOINSTALL) $(TOPLEVEL_PKG)/$*
$(IREF_LIST): %_iref:
	$(GODEP) $(TOPLEVEL_PKG)/$*
$(TEST_LIST): %_test:
	$(GOTEST) $(TOPLEVEL_PKG)/$*
$(LINT_LIST): %_lint:
	$(GOLINT) $(TOPLEVEL_PKG)/$*
$(FMT_TEST): %_fmt:
	$(GOFMT) ./$*
