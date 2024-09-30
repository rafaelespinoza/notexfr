#!/usr/bin/env -S just -f

BIN := justfile_directory() / "bin" / "notexfr"

GO := "go"
GOSEC := "gosec"
PKG_IMPORT_PATH := "github.com/rafaelespinoza/notexfr"
_CORE_SRC_PKG_PATHS := PKG_IMPORT_PATH + " " + PKG_IMPORT_PATH / "internal" / "..."
_LDFLAGS_BASE_PREFIX := "-X " + PKG_IMPORT_PATH + "/internal/version"
_LDFLAGS_DELIMITER := "\n\t"
LDFLAGS := (
  _LDFLAGS_BASE_PREFIX + ".BranchName=" + `git rev-parse --abbrev-ref HEAD` + _LDFLAGS_DELIMITER +
  _LDFLAGS_BASE_PREFIX + ".BuildTime=" + `date --utc +%FT%T%z` + _LDFLAGS_DELIMITER +
  _LDFLAGS_BASE_PREFIX + ".CommitHash=" + `git rev-parse --short=7 HEAD` + _LDFLAGS_DELIMITER +
  _LDFLAGS_BASE_PREFIX + ".GoOSArch=" + `go version | awk '{ print $4 }' | tr '/' '_'` + _LDFLAGS_DELIMITER +
  _LDFLAGS_BASE_PREFIX + ".GoVersion=" + `go version | awk '{ print $3 }'` + _LDFLAGS_DELIMITER +
  _LDFLAGS_BASE_PREFIX + ".ReleaseTag=" + `git describe --tag 2>/dev/null || echo 'dev'` + _LDFLAGS_DELIMITER
)
PKG_PATH := "./..."

# list available recipes
default:
    @just --list --unsorted

# compile a binary to directory bin/
build:
    #!/bin/sh
    set -eu
    mkdir -pv {{ parent_directory(BIN) }}
    {{ GO }} build -o="{{ BIN }}" -v -ldflags="{{ LDFLAGS }}"
    {{ BIN }} version

alias b := build

# get module dependencies, tidy them up, vendor them
mod:
    {{ GO }} mod tidy && {{ GO }} mod vendor

# run the tests
test *ARGS='':
    {{ GO }} test {{ ARGS }} {{ PKG_PATH }}

# This Justfile won't install the scanner binary for you, so check out the gosec
# README for instructions: https://github.com/securego/gosec
# If necessary, specify the path to the built binary with the GOSEC env var.
#
# run a security scanner over the source code
gosec *ARGS='':
    {{ GOSEC }} {{ ARGS }} {{ PKG_PATH }}

# examine source code for suspicious constructs
vet *ARGS='':
    {{ GO }} vet {{ ARGS }} {{ PKG_PATH }}

# execute the compiled binary
run *ARGS='':
    {{ BIN }} {{ ARGS }}

alias r := run

# compile a binary and run it
buildrun *ARGS='': build (run ARGS)

alias br := buildrun
