# Dependency bump and Protobuf edition 2026

## Purpose

Bring the module, the Taskfile toolchain, and the repo's `.proto` files up to the current line in one change. Every proto this repo owns uses `edition = "2026"` and the Opaque Go API that edition defaults to. The plugin keeps accepting older syntax as input.

## Decisions

- One change. Tool pins and the edition move land together.
- `go.mod` says `go 1.27.0`. There is no `toolchain` line.
- `google.golang.org/protobuf` is `v1.37.0-rc.1`. That is the first release whose `protoc-gen-go` declares edition 2026 support. Stable `v1.36.12` does not. If tag `v1.37.0` exists and is not a pre-release when the change is made, use `v1.37.0` instead. Do not float to a later release candidate.
- `protoc-gen-go` is installed from that same module version into `.tool`. It is not a second pin in `.binny.yaml`.
- Buf is at least `v1.75.0`, the release that parses edition 2026 and records `import option` dependencies in images.
- The Makefile bootstrap installs the binny version named by `BINNY_VERSION` in `Taskfile.yaml`. The curl installer at `get.anchore.io` is removed.
- CI `go-version` stays `stable`. The golangci-lint action stays `golangci/golangci-lint-action@v9`. The linter version input becomes the full tag pinned in `.binny.yaml`.
- Files this repo owns use `edition = "2026"` with no `features.(pb.go).api_level` override. The edition default is `API_OPAQUE`.
- The plugin reports `SupportedEditionsMinimum = EDITION_PROTO2` and `SupportedEditionsMaximum = EDITION_2026`.
- proto2, proto3, edition 2023, and edition 2024 remain supported inputs. The README table lists them, and lists edition 2026 as the recommended syntax.
- Field numbers and types on `fieldmeta.v1` extensions stay: `log` 50100, `sensitive` 50101, `immutable` 50102, `db` 50103, `mask` 50104, `tags` 50200.
- Published extension names stay those six strings.

## Toolchain

### Versions

Surveyed 2026-10-22. At implementation time, take a newer stable tag when one exists. Buf stays on `v1.75.0` or newer. The protobuf module follows the rule in Decisions, not this table.

| Pin | Now | Target |
| --- | --- | --- |
| `go` directive | 1.26.0 | 1.27.0 |
| `google.golang.org/protobuf` | v1.36.11 | v1.37.0-rc.1, or stable v1.37.0 if that tag exists |
| task | v3.49.1 | v3.54.0 |
| golangci-lint | v2.11.4 | v2.15.0 |
| buf | v1.67.0 | v1.75.0 |
| binny (`BINNY_VERSION`) | v0.13.0 | v0.13.3 |

Raise `BINNY_VERSION` first, then run `task update-tools` so the update uses that binny. If `binny update` leaves a pin behind the latest stable tag, set the pin to that tag by hand.

`.github/workflows/lint.yml` sets `version` to the same full golangci-lint tag `.binny.yaml` records (for example `v2.15.0`, not `v2.15`).

### Makefile

`BINNY_VERSION` lives only in `Taskfile.yaml`. The Makefile reads it:

```make
BINNY_VERSION := $(shell awk '/^  BINNY_VERSION:/ {print $$2}' Taskfile.yaml)
```

The `$(BINNY)` recipe is:

```make
mkdir -p $(TOOL_DIR)
GOBIN=$(abspath $(TOOL_DIR)) go install github.com/anchore/binny/cmd/binny@$(BINNY_VERSION)
```

`make` requires Go on `PATH`, which the rest of the build already requires.

### protoc-gen-go

Add an internal Taskfile task that installs the plugin next to the other tools:

```
GOBIN=$(pwd)/.tool go install google.golang.org/protobuf/cmd/protoc-gen-go@$(go list -m -f '{{.Version}}' google.golang.org/protobuf)
```

The task's `generates` entry is `.tool/protoc-gen-go`. Its `status` check fails when the binary is missing or when `protoc-gen-go --version` does not report the module version, so a module bump rebuilds the binary.

`proto-generate` depends on `tools` and this task. It runs buf with `PATH=.tool:$PATH`.

`proto-generate-fieldmeta` already puts `.tool` first on `PATH` and does not invoke `protoc-gen-go`. It does not gain this dependency.

## Protobuf edition 2026

Set `edition = "2026"` on:

- `fieldmeta/v1/options.proto`
- `testdata/testpb/basic.proto`
- `testdata/testpb/oneof.proto`
- `testdata/testpb/no_options.proto`
- `testdata/testpb/tags_compat.proto`
- the new `testdata/testpb/import_option.proto`

No file sets `features.(pb.go).api_level`.

### Feature defaults this repo accepts

Edition 2026 changes these defaults relative to the edition 2023 files here today. This change takes the new defaults.

| Feature | Edition 2026 default | What this repo does |
| --- | --- | --- |
| `(pb.go).api_level` | `API_OPAQUE` | Regenerate. Rewrite tests that assign exported fields. |
| `default_symbol_visibility` | `STRICT` | Leave the default. See visibility below. |
| `enforce_naming_style` | `STYLE2026` | Keep current names. They are snake_case and are not `descriptor` or a `has_` / `set_` / `get_` / `clear_` / `_value` collision. |
| `enforce_proto_limits` | `PROTO_LIMITS2026` | No override. Messages are small. |

Wire-facing defaults used by these files stay `field_presence = EXPLICIT`, packed repeated fields, and UTF-8 verify. `options.proto` has no messages, enums, or oneofs.

`cc_enable_arenas` is removed in edition 2026. This repo does not set it, and `buf.yaml` does not use managed mode.

### Visibility

`STRICT` makes top-level messages and enums local unless they are marked `export`. Nested types cannot be marked `export`. `export` and `local` apply to messages and enums.

`Outer.Inner` in `basic.proto` is used only inside that file. It stays nested and unmarked.

Testdata messages are not imported by other `.proto` files. They stay unmarked.

`options.proto` contains only extensions of `google.protobuf.FieldOptions`. It gets no visibility override on the first attempt. Testdata files and the new fixture must still resolve `(fieldmeta.v1.log)`, `(fieldmeta.v1.sensitive)`, `(fieldmeta.v1.immutable)`, `(fieldmeta.v1.db)`, `(fieldmeta.v1.mask)`, and `(fieldmeta.v1.tags)`.

If `buf lint` or `buf generate` rejects those references because `STRICT` hides the extensions, add this single line to `options.proto` and nowhere else:

```protobuf
option features.default_symbol_visibility = EXPORT_TOP_LEVEL;
```

That exception exists so the published custom options stay importable. Testdata files do not get it.

### import option fixture

`testdata/testpb/import_option.proto`:

- `edition = "2026"`
- `package testpb`
- the same `go_package` as the other files in that directory
- `import option "fieldmeta/v1/options.proto";`
- no regular `import` of that file
- one message with one string field whose `(fieldmeta.v1.log)` is a non-empty string

`buf.gen.fieldmeta.yaml` already uses input directory `.`. The generate task passes `--path testdata/testpb/`, which includes the new file.

## Plugin

In `cmd/protoc-gen-go-fieldmeta/main.go`, set `gen.SupportedEditionsMaximum` to `descriptorpb.Edition_EDITION_2026`. Leave `SupportedEditionsMinimum` at `descriptorpb.Edition_EDITION_PROTO2`. Leave `SupportedFeatures` as it is.

The generator keeps reading field options through `protogen` and emitting helpers that use `protoreflect`. It does not branch on edition.

If the `import option` fixture's generated helper does not see the log option, fix the read path so an option the compiler placed on the field is emitted. Do not add an edition switch for that fix.

## Tests

`testdata/testpb/fieldmeta_test.go` and `fieldmetautil/reflect_test.go` construct Open structs (`&User{Email: proto.String(...)}`, exported oneof wrappers, `u.Email = proto.String(...)`). Rewrite those constructions and reads to the Opaque setters and getters `protoc-gen-go` emits for edition 2026, including each oneof arm. Keep the nil and zero-value cases.

Helpers such as `LogFields_User` already use `msg.ProtoReflect()`. They stay unless regeneration shows they no longer compile.

Add a test next to `TestLogFields_User` for the import-option message. It builds the message with Opaque setters, calls the generated log helper, and checks the configured log key and value.

Regenerate and commit:

- `fieldmeta/v1/options.pb.go`
- `testdata/testpb/*.pb.go`
- `testdata/testpb/*_fieldmeta.pb.go`, including one for the new proto

After `task proto-generate` and `task proto-generate-fieldmeta`, `git status` shows no changes under those generated files.

## Documentation

- `README.md` quick-start sample uses `edition = "2026"`. The compatibility table keeps proto2, proto3, and edition 2023, and adds edition 2024 and edition 2026. The surrounding prose calls edition 2026 the syntax to use for new files.
- `MIGRATION.md` adds a row: edition 2023 to edition 2026. The steps are to change the edition string, switch call sites to Opaque accessors, and mark `export` on any top-level message another `.proto` file imports. The row states that `fieldmeta/v1` options are extensions, so consumers do not mark a message `export` in that file.
- `CLAUDE.md` says Go 1.27+.

## CI

Add `.github/workflows/proto.yml`. It runs on push to `main` and on pull requests. Permissions are `contents: read`. One job, `ubuntu-latest`:

1. `actions/checkout@v6` with `fetch-depth: 0`.
2. `git fetch origin main:refs/heads/main`, so `.git#branch=main` resolves on pull requests. Checkout's current ref is the PR head, and `branch=main` reads `refs/heads/main`, not `refs/remotes/origin/main`.
3. `actions/setup-go@v6` with `go-version: stable`.
4. `make proto-lint`
5. `make proto-breaking`

The job uses the Makefile binny pin and `.binny.yaml`. It does not install buf from a GitHub Action. It does not regenerate code.

`.github/workflows/test.yml` is unchanged. `.github/workflows/lint.yml` changes only the golangci-lint `version` value.

There is no `.golangci.yml` today. If golangci-lint `v2.15` (or the newer pin) reports findings, fix them in this change. Add a config file only to disable a single linter that is a false positive, with a comment that names the finding and why it is ignored.

## Gates

All of these pass before the change is done:

- `task test` (`go test -race ./...`)
- `task lint`
- `task proto-lint`
- `task proto-breaking`
- `task proto-generate` and `task proto-generate-fieldmeta` leave a clean generated tree
- the new proto workflow's two `make` targets are the same commands as the local proto gates

`buf breaking` stays green for `fieldmeta/v1/options.proto`. If the only breaking failures are under `testdata/`, extend `buf.yaml` to:

```yaml
breaking:
  use:
    - FILE
  ignore:
    - testdata
```

Lint already ignores that path. Do not add `fieldmeta/v1` to `breaking.ignore`.

## Out of scope

- A plugin release, a buf schema push, or a module tag.
- Dependabot config changes, and bumps of `actions/checkout`, `actions/setup-go`, or `golangci/golangci-lint-action` major versions.
- Dropping proto2, proto3, edition 2023, or edition 2024 as inputs the plugin accepts.
- Pinning `API_OPEN` or `API_HYBRID`.
- Renaming `fieldmeta.v1` extensions.
- New generator behavior beyond the edition ceiling and honoring options that `import option` already attaches to fields.
- An OS matrix for the proto workflow.

## Work sequence

1. Raise `BINNY_VERSION`, point the Makefile at it, run `task update-tools`, and copy the golangci-lint tag into `lint.yml`.
2. Set `go 1.27.0`, require the protobuf module version from Decisions, and `go mod tidy`.
3. Add the `protoc-gen-go` Taskfile task and point `proto-generate` at `.tool`.
4. Raise `SupportedEditionsMaximum` to `EDITION_2026`.
5. Switch existing protos to edition 2026, add `import_option.proto`, regenerate, and rewrite tests to Opaque accessors.
6. Apply the `options.proto` visibility exception only if extension references fail.
7. Update `README.md`, `MIGRATION.md`, and `CLAUDE.md`.
8. Add `.github/workflows/proto.yml`.
9. Run the gates. If `buf breaking` fails only on testdata, ignore that path.
