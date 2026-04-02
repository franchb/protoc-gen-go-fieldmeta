# Migration from @gotags / protoc-gen-go-struct-tag

## Why Migrate

The Protobuf Go Opaque API removes direct struct access, breaking struct-tag injection tools. `protoc-gen-go-fieldmeta` replaces them with first-class protobuf extensions.

## Quick Migration (Tags Shim)

The fastest path: replace `@gotags` comments with `fieldmeta.v1.tags`:

**Before (with @gotags):**
```protobuf
message User {
  string email = 1; // @gotags: validate:"required,email" yaml:"email"
}
```

**After (with fieldmeta.v1.tags):**
```protobuf
import "fieldmeta/v1/options.proto";

message User {
  string email = 1 [
    (fieldmeta.v1.tags) = "validate:\"required,email\" yaml:\"email\""
  ];
}
```

**Usage changes:**
```go
// Before: reflect-based struct tag reading
// tag := reflect.TypeOf(User{}).Field(0).Tag.Get("validate")

// After: generated helper
tag := GetFieldTag_User("email", "validate") // "required,email"
```

## Recommended Migration (Structured Options)

For new code or a full migration, use structured options instead of raw tags:

```protobuf
message User {
  string email = 1 [
    (fieldmeta.v1.log)       = "user_email",
    (fieldmeta.v1.db)        = "email_address",
    (fieldmeta.v1.sensitive) = true
  ];
}
```

This gives you type-safe, purpose-built helpers instead of string parsing.

## Build Pipeline Changes

**Before:**
```yaml
# buf.gen.yaml
plugins:
  - local: protoc-gen-go
    out: gen/go
    opt: paths=source_relative
# + post-processing step to inject struct tags
```

**After:**
```yaml
# buf.gen.yaml
plugins:
  - local: protoc-gen-go
    out: gen/go
    opt: paths=source_relative
  - local: protoc-gen-go-fieldmeta
    out: gen/go
    opt: paths=source_relative
```

No post-processing step needed.

## Edition Migration Path

| Current | Target | Steps |
|---------|--------|-------|
| proto3 + @gotags | proto3 + fieldmeta.tags | Replace comments with `(fieldmeta.v1.tags)` extensions |
| proto3 + fieldmeta.tags | proto3 + structured | Replace `tags` with `log`, `db`, `sensitive`, etc. |
| proto3 + structured | edition 2023 + structured | Change `syntax = "proto3"` to `edition = "2023"`, remove `optional` keywords |
| edition 2023 + structured | Opaque API | Enable opaque API — generated helpers already work |

Each step is independent. You can stop at any point and have a working setup.
