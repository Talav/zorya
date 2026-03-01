# File Uploads

Zorya handles multipart file uploads via `body:"multipart"` using the talav/schema library. File fields use `*multipart.FileHeader` (or `[]*multipart.FileHeader` for multiple files), which exposes filename, size, and an `Open()` method to read the file content.

## Input struct

Use `body:"multipart"` with a Body struct. The schema library supports `query`, `path`, `header`, and `cookie` locations — not `form`. Multipart form fields go in the body:

```go
import "mime/multipart"

type UploadInput struct {
    Body struct {
        // Regular form fields
        Description string `schema:"description"`

        // Single file upload (*multipart.FileHeader from talav/schema)
        File *multipart.FileHeader `schema:"file"`

        // Multiple file upload
        Attachments []*multipart.FileHeader `schema:"attachments"`
    } `body:"multipart"`
}
```

## Handler

`*multipart.FileHeader` exposes `Filename`, `Size`, and `Open()` for reading the stream:

```go
func uploadFile(ctx context.Context, input *UploadInput) (*UploadOutput, error) {
    if input.Body.File == nil {
        return nil, zorya.Error400BadRequest("file is required")
    }

    if input.Body.File.Size > maxFileSize {
        return nil, zorya.Error400BadRequest(
            fmt.Sprintf("file too large: %d bytes (max %d bytes)", input.Body.File.Size, maxFileSize),
        )
    }

    f, err := input.Body.File.Open()
    if err != nil {
        return nil, zorya.Error500InternalServerError("could not open uploaded file", err)
    }
    defer f.Close()

    buf := make([]byte, 512)
    n, err := f.Read(buf)
    if err != nil && err != io.EOF {
        return nil, zorya.Error500InternalServerError("could not read uploaded file", err)
    }
    contentType := http.DetectContentType(buf[:n])

    if !strings.HasPrefix(contentType, "image/") {
        return nil, zorya.Error415UnsupportedMediaType(
            fmt.Sprintf("only image uploads are accepted, got: %s", contentType),
        )
    }

    out := &UploadOutput{Status: http.StatusCreated}
    out.Body.Filename = input.Body.File.Filename
    out.Body.Size = input.Body.File.Size
    out.Body.ContentType = contentType
    out.Body.Description = input.Body.Description
    return out, nil
}
```

## Output

```go
type UploadOutput struct {
    Status int `json:"-"`
    Body   struct {
        Filename    string `json:"filename"`
        Size        int64  `json:"size"`
        ContentType string `json:"content_type"`
        Description string `json:"description,omitempty"`
    } `body:"structured"`
}
```

## Body size limit

The default body size limit is 1 MB. Increase it for file upload routes:

```go
zorya.Post(api, "/uploads", uploadFile, func(r *zorya.BaseRoute) {
    r.MaxBodyBytes = 50 * 1024 * 1024 // 50 MB
})
```

Set to `-1` to remove the limit entirely (not recommended for production).

## Multiple files

When the client uploads multiple files under the same field name, use `[]*multipart.FileHeader`:

```go
type BulkUploadInput struct {
    Body struct {
        Files []*multipart.FileHeader `schema:"files"`
    } `body:"multipart"`
}

func bulkUpload(ctx context.Context, input *BulkUploadInput) (*BulkOutput, error) {
    for _, fh := range input.Body.Files {
        f, err := fh.Open()
        if err != nil {
            return nil, zorya.Error500InternalServerError("could not open file", err)
        }
        defer f.Close()
        // process each file...
    }
    // ...
}
```

## Client example (curl)

```bash
# Single file
curl -X POST http://localhost:8080/uploads \
     -F "file=@photo.jpg"

# Multiple files
curl -X POST http://localhost:8080/uploads \
     -F "files=@doc1.pdf" \
     -F "files=@doc2.pdf"
```
