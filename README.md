# protoc-gen-go-fieldmeta

A protoc plugin that generates typed Go helper functions for reading Protocol Buffer field metadata options.

## Why

The Protobuf Go Opaque API (introduced in `google.golang.org/protobuf` v1.36+) removes direct struct field access, breaking tools like `protoc-gen-go-struct-tag` / `@gotags` that rely on injecting struct tags via post-processing. `protoc-gen-go-fieldmeta` solves this by reading custom extensions from `FieldOptions` at generation time and emitting standalone helper functions that use only `protoreflect` — no `reflect` package, no struct tag hacks.

## Install

```bash
go install github.com/franchb/protoc-gen-go-fieldmeta/cmd/protoc-gen-go-fieldmeta@latest
```

## Getting the proto import

Your `.proto` files need `import "fieldmeta/v1/options.proto"`. There are several ways to make this import available depending on your toolchain.

### Option A: Buf + BSR (recommended)

Add the dependency to your project's `buf.yaml`:

```yaml
# buf.yaml
version: v2
modules:
  - path: proto
deps:
  - buf.build/franchb-oss/protoc-gen-go-fieldmeta
```

Then run:

```bash
buf dep update
```

The import resolves automatically when you run `buf generate`.

### Option B: Buf + local copy

Clone or copy `fieldmeta/v1/options.proto` into your project, then include its parent in `buf.yaml`:

```yaml
# buf.yaml
version: v2
modules:
  - path: proto
  - path: third_party
```

Where `third_party/fieldmeta/v1/options.proto` contains the file.

### Option C: protoc + BSR (via buf export)

Export the proto from BSR, then use it with protoc:

```bash
buf export buf.build/franchb-oss/protoc-gen-go-fieldmeta -o ./third_party

protoc -I. -I./third_party \
  --go_out=. --go_opt=paths=source_relative \
  --go-fieldmeta_out=. --go-fieldmeta_opt=paths=source_relative \
  your/service.proto
```

### Option D: protoc + Go module (no BSR)

The proto file ships inside the Go module. Point protoc at it:

```bash
go install github.com/franchb/protoc-gen-go-fieldmeta/cmd/protoc-gen-go-fieldmeta@latest

FIELDMETA_INCLUDE=$(go list -m -f '{{.Dir}}' github.com/franchb/protoc-gen-go-fieldmeta)

protoc -I. -I"$FIELDMETA_INCLUDE" \
  --go_out=. --go_opt=paths=source_relative \
  --go-fieldmeta_out=. --go-fieldmeta_opt=paths=source_relative \
  your/service.proto
```

## Quick Start

**1. Define options in your .proto file:**

```protobuf
edition = "2023";

import "fieldmeta/v1/options.proto";

message User {
  string email = 1 [
    (fieldmeta.v1.log) = "user_email",
    (fieldmeta.v1.db)  = "email_address"
  ];
  string password_hash = 2 [
    (fieldmeta.v1.sensitive) = true
  ];
}
```

**2. Add the plugin to your `buf.gen.yaml`:**

```yaml
version: v2
plugins:
  - local: protoc-gen-go
    out: gen/go
    opt: paths=source_relative
  - local: protoc-gen-go-fieldmeta
    out: gen/go
    opt: paths=source_relative
```

**3. Generate and use:**

```bash
buf generate
```

```go
fields := LogFields_User(msg)        // map["user_email"] = "alice@example.com"
names  := SensitiveFieldNames_User() // ["password_hash"]
safe   := RedactSensitive_User(msg)  // deep copy with password_hash cleared
cols   := FieldDbColumns_User()      // map["email"] = "email_address"
```

## Option Reference

| Extension | Type | Description |
|-----------|------|-------------|
| `fieldmeta.v1.log` | `string` | Structured logging key |
| `fieldmeta.v1.sensitive` | `bool` | Marks field as containing sensitive data |
| `fieldmeta.v1.immutable` | `bool` | Marks field as immutable after creation |
| `fieldmeta.v1.db` | `string` | Database column name mapping |
| `fieldmeta.v1.mask` | `string` | Masking behavior hint (`"email"`, `"phone"`, `"full"`) |
| `fieldmeta.v1.tags` | `string` | Raw struct tag string (migration shim from `@gotags`) |

## Generated Functions

For each annotated message `Msg`, the plugin generates:

| Function | Generated When | Signature |
|----------|---------------|-----------|
| `LogFields_Msg` | Any field has `log` | `(msg *Msg) map[string]any` |
| `SensitiveFieldNames_Msg` | Any field has `sensitive` | `() []string` |
| `RedactSensitive_Msg` | Any field has `sensitive` | `(msg *Msg) *Msg` |
| `FieldDbColumns_Msg` | Any field has `db` | `() map[string]string` |
| `FieldTags_Msg` | Any field has `tags` | `() map[string]string` |
| `GetFieldTag_Msg` | Any field has `tags` | `(protoFieldName, tagKey string) string` |

Messages with no fieldmeta annotations produce no generated file.

## Runtime Library

The `fieldmetautil` package provides generic, reflection-based access to fieldmeta options on any `proto.Message`:

```go
import "github.com/franchb/protoc-gen-go-fieldmeta/fieldmetautil"

fields := fieldmetautil.LogFields(anyMsg)
names  := fieldmetautil.SensitiveFieldNames(anyMsg)
safe   := fieldmetautil.RedactSensitive(anyMsg)
val    := fieldmetautil.GetTagValue(fieldDescriptor, "validate")
```

Use the generated helpers for type safety and zero reflection overhead. Use `fieldmetautil` for generic middleware or when you don't know the message type at compile time.

## Migration from @gotags

See [MIGRATION.md](MIGRATION.md) for a step-by-step guide.

Quick version: use `fieldmeta.v1.tags` as a drop-in shim, then migrate to structured options.

## Edition Compatibility

| Proto Syntax | Supported |
|-------------|-----------|
| `proto2` | Yes |
| `proto3` | Yes |
| `edition = "2023"` | Yes |

## Contributing

```bash
make          # bootstrap tools + lint + test
task test     # run tests
task lint     # run linter
```

## License

Apache 2.0
