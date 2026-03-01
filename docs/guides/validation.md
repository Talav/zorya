# Validation

Zorya validates request inputs automatically before the handler is called. Validation errors are returned as RFC 9457 structured error responses.

## Automatic tag-based validation

Add `validate` tags to your input struct fields. Zorya reads them and runs validation on the decoded input.

```go
type CreateUserInput struct {
    Body struct {
        Name     string `json:"name"     validate:"required,min=2,max=100"`
        Email    string `json:"email"    validate:"required,email"`
        Age      int    `json:"age"      validate:"gte=0,lte=150"`
        Role     string `json:"role"     validate:"oneof=admin member viewer"`
        Password string `json:"password" validate:"required,min=8"`
    } `body:"structured"`
}
```

If validation fails, the handler is never called and Zorya returns `422 Unprocessable Entity` with an `ErrorModel` body listing each failed field.

For the full list of constraint keywords, see the [go-playground/validator documentation](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Baked_In_Validators_and_Tags).

## Plugging in a validator

By default Zorya uses go-playground/validator. Supply your own by implementing `Validator`:

```go
type Validator interface {
    Validate(ctx context.Context, v any) []error
}
```

Pass it when creating the API:

```go
api := zorya.NewAPI(adapter, zorya.WithValidator(myValidator))
```

Each returned error should implement `ErrorDetailer` for field-level information in the response:

```go
type ErrorDetailer interface {
    ErrorDetail() *ErrorDetail
}
```

## Example: go-playground/validator integration

The following shows a complete custom validator that produces structured `ErrorDetail` objects:

```go
import (
    "context"
    "fmt"

    "github.com/go-playground/validator/v10"
    "github.com/talav/zorya"
)

type PlaygroundValidator struct {
    v *validator.Validate
}

func NewPlaygroundValidator() *PlaygroundValidator {
    return &PlaygroundValidator{v: validator.New()}
}

func (pv *PlaygroundValidator) Validate(ctx context.Context, input any, _ *schema.StructMetadata) []error {
    err := pv.v.StructCtx(ctx, input)
    if err == nil {
        return nil
    }

    var errs []error
    for _, fe := range err.(validator.ValidationErrors) {
        errs = append(errs, &zorya.ErrorDetail{
            Code:     fe.Tag(),
            Message:  fmt.Sprintf("field validation failed on '%s' constraint", fe.Tag()),
            Location: "body." + fe.Field(),
        })
    }
    return errs
}

// Wire up:
api := zorya.NewAPI(adapter, zorya.WithValidator(NewPlaygroundValidator()))
```

## Validation error response shape

When validation fails, clients receive:

```json
{
  "title": "Unprocessable Entity",
  "status": 422,
  "errors": [
    {
      "code": "required",
      "message": "field validation failed on 'required' constraint",
      "location": "body.Email"
    }
  ]
}
```
