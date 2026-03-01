# Conditional Requests

Zorya provides first-class support for HTTP conditional requests: `If-Match`, `If-None-Match`, `If-Modified-Since`, and `If-Unmodified-Since`. Embed `conditional.Params` in your input struct and call `CheckPreconditions` in the handler.

## Import

```go
import "github.com/talav/zorya/conditional"
```

## Embedding Params

```go
type GetUserInput struct {
    ID string `schema:"id,location=path"`
    conditional.Params  // adds If-Match, If-None-Match, If-Modified-Since, If-Unmodified-Since
}
```

`Params` fields:

```go
type Params struct {
    IfMatch           []string  `schema:"If-Match,location=header"`
    IfNoneMatch       []string  `schema:"If-None-Match,location=header"`
    IfModifiedSince   time.Time `schema:"If-Modified-Since,location=header"`
    IfUnmodifiedSince time.Time `schema:"If-Unmodified-Since,location=header"`
}
```

## Checking preconditions

Call `CheckPreconditions` early in the handler, before loading data:

```go
func getUser(ctx context.Context, input *GetUserInput) (*UserOutput, error) {
    user := db.Get(input.ID)
    if user == nil {
        return nil, zorya.Error404NotFound("user not found")
    }

    // Check If-None-Match / If-Modified-Since (read operation: isWrite=false)
    if err := input.Params.CheckPreconditions(user.ETag, user.UpdatedAt, false); err != nil {
        return nil, err  // returns 304 Not Modified or 412 Precondition Failed
    }

    out := &UserOutput{}
    out.ETag         = user.ETag
    out.LastModified = user.UpdatedAt.Format(http.TimeFormat)
    out.Body         = user
    return out, nil
}
```

The third argument `isWrite` controls semantics:

- `false` (GET/HEAD): `If-None-Match` triggers `304 Not Modified`
- `true` (PUT/DELETE): `If-None-Match: *` triggers `412 Precondition Failed`

## Write (optimistic locking)

For updates, require the client to send a matching ETag:

```go
type UpdateUserInput struct {
    ID   string `schema:"id,location=path"`
    conditional.Params
    Body struct {
        Name string `json:"name" validate:"required"`
    } `body:"structured"`
}

func updateUser(ctx context.Context, input *UpdateUserInput) (*UserOutput, error) {
    user := db.Get(input.ID)
    if user == nil {
        return nil, zorya.Error404NotFound("user not found")
    }

    // isWrite=true: If-Match must match; If-None-Match: * prevents update if exists
    if err := input.Params.CheckPreconditions(user.ETag, user.UpdatedAt, true); err != nil {
        return nil, err
    }

    user.Name = input.Body.Name
    user.ETag = newETag(user)
    db.Save(user)

    return &UserOutput{Body: user, ETag: user.ETag}, nil
}
```

## ETag generation

Any stable hash of the resource state works as an ETag. A simple approach:

```go
import (
    "crypto/sha256"
    "fmt"
    "encoding/json"
)

func newETag(v any) string {
    b, _ := json.Marshal(v)
    return fmt.Sprintf(`"%x"`, sha256.Sum256(b))
}
```

## CheckPreconditions signature

```go
func (p *Params) CheckPreconditions(
    currentETag     string,
    currentModified time.Time,
    isWrite         bool,
) zorya.StatusError
```

Returns `nil` if preconditions pass, or a `StatusError` (304 or 412) otherwise.
