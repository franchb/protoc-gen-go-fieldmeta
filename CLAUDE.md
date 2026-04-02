# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

`protoc-gen-go-fieldmeta` is a protoc plugin that generates Go code with field metadata from Protocol Buffer definitions. Go 1.26+, Apache 2.0 license.

## Commands

`make` bootstraps everything (installs binny via curl, then task via binny). All targets shim through to go-task, so `make lint` and `task lint` are equivalent. Dev tools (golangci-lint, buf) are managed by [binny](https://github.com/anchore/binny) and installed into `.tool/`.

```sh
make                  # bootstrap + default (lint + test)
make tools            # install all dev tools
task build            # go build ./...
task test             # go test -race ./...
task lint             # golangci-lint run ./...
task fmt              # go fmt ./...
task proto-generate   # buf generate
task proto-lint       # buf lint
task proto-format     # buf format -w
task proto-breaking   # buf breaking --against main
task update-tools     # update pinned tool versions in .binny.yaml
task clean            # remove .tool/
```

Run a single test: `go test -race -run TestName ./path/to/pkg`

## Tool versions

Pinned in `.binny.yaml`, installed to `.tool/` (gitignored). Reference tools as `.tool/<name>` in any new Taskfile targets.
