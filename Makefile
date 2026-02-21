# Makefile for gode repository
# Provides targets for building in normal, debug and release modes, and
# running the full test suite.

# default target
all: build

.PHONY: all build debug release test clean

# standard build; output binary named "gode" in the workspace root
# build only the main package (the root of the module) rather than
# attempting to write a single binary for every package.
build:
	@echo "building gode"
	go build -o gode .

# debug build disables inlining and optimizations so you can attach a debugger
debug:
	@echo "building debug binary (no optimizations)"
	go build -gcflags "all=-N -l" -o gode-debug .

# release build strips symbol table and disables DWARF info for smaller size
release:
	@echo "building release binary"
	go build -trimpath -ldflags "-s -w" -o gode .

# run every test under the module tree
test:
	@echo "running tests"
	go test ./...

# clean up build artifacts
clean:
	@echo "cleaning"
	go clean ./...
	rm -f gode gode-debug
