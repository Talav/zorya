# File Uploads

Zorya provides built-in support for file uploads using multipart/form-data encoding with automatic OpenAPI documentation generation.

## Overview

File uploads are handled using the `body:"multipart"` tag on input structs. Zorya automatically:

- Parses multipart/form-data requests
- Handles binary file content
- Generates correct OpenAPI schemas with `format: binary`
- Sets proper Content-Type for file fields

## Basic File Upload

### Input Struct

```go
type UploadFileInput struct {
    // Optional path/query parameters
    ResourceID string `schema:"resource_id,location=path,required=true"`
    Overwrite  bool   `schema:"overwrite,location=query"`
    
    // Multipart body with file
    Body struct {
        // File field - use []byte for binary content
        File        []byte `json:"file" openapi:"format=binary,description=File content"`
        
        // Additional form fields
        Filename    string `json:"filename" openapi:"description=Name of the file"`
        Description string `json:"description" openapi:"description=Optional file description"`
    } `body:"multipart,required"`
}
```

### Output Struct

```go
type UploadFileOutput struct {
    Status int `status:"201"`
    Body struct {
        ID       int64  `json:"id"`
        Filename string `json:"filename"`
        Size     int64  `json:"size"`
        URL      string `json:"url"`
    } `body:"structured"`
}
```

### Handler

```go
func uploadFile(ctx context.Context, input *UploadFileInput) (*UploadFileOutput, error) {
    // Access file content
    fileContent := input.Body.File
    filename := input.Body.Filename
    
    // Save file (example)
    fileID, err := fileService.Save(ctx, filename, fileContent)
    if err != nil {
        return nil, zorya.Error500InternalServerError("Failed to save file", err)
    }
    
    return &UploadFileOutput{
        Status: http.StatusCreated,
        Body: struct {
            ID       int64  `json:"id"`
            Filename string `json:"filename"`
            Size     int64  `json:"size"`
            URL      string `json:"url"`
        }{
            ID:       fileID,
            Filename: filename,
            Size:     int64(len(fileContent)),
            URL:      fmt.Sprintf("/files/%d", fileID),
        },
    }, nil
}
```

### Route Registration

```go
zorya.Post(api, "/resources/{resource_id}/files", uploadFile)
```

## Multiple File Upload

You can upload multiple files by using a slice of byte slices:

```go
type UploadMultipleFilesInput struct {
    Body struct {
        // Multiple files
        Files [][]byte `json:"files" openapi:"format=binary,description=Multiple files"`
        
        // Metadata for each file
        Filenames    []string `json:"filenames"`
        Descriptions []string `json:"descriptions"`
    } `body:"multipart,required"`
}
```

## File Upload with Metadata

Combine file upload with structured metadata:

```go
type UploadWithMetadataInput struct {
    Body struct {
        // File
        File []byte `json:"file" openapi:"format=binary"`
        
        // Structured metadata
        Metadata struct {
            Title       string   `json:"title"`
            Tags        []string `json:"tags"`
            Category    string   `json:"category"`
            IsPublic    bool     `json:"is_public"`
        } `json:"metadata"`
    } `body:"multipart,required"`
}
```

## Size Limits

Set body size limits for file uploads to prevent abuse:

```go
zorya.Post(api, "/files", uploadFile,
    func(route *zorya.BaseRoute) {
        // Allow up to 10MB
        route.MaxBodyBytes = 10 * 1024 * 1024
        
        // Increase read timeout for large files
        route.BodyReadTimeout = 30 * time.Second
    },
)
```

## File Types and Validation

### Validate File Type

```go
func uploadFile(ctx context.Context, input *UploadFileInput) (*UploadFileOutput, error) {
    // Detect file type
    contentType := http.DetectContentType(input.Body.File)
    
    // Validate allowed types
    allowedTypes := map[string]bool{
        "image/jpeg": true,
        "image/png":  true,
        "image/gif":  true,
        "application/pdf": true,
    }
    
    if !allowedTypes[contentType] {
        return nil, zorya.Error400BadRequest("Invalid file type")
    }
    
    // Process file...
}
```

### Validate File Size

```go
func uploadFile(ctx context.Context, input *UploadFileInput) (*UploadFileOutput, error) {
    maxSize := 5 * 1024 * 1024 // 5MB
    if len(input.Body.File) > maxSize {
        return nil, zorya.Error400BadRequest("File too large")
    }
    
    // Process file...
}
```

### Validate File Extension

```go
func uploadFile(ctx context.Context, input *UploadFileInput) (*UploadFileOutput, error) {
    ext := filepath.Ext(input.Body.Filename)
    allowedExts := map[string]bool{
        ".jpg":  true,
        ".jpeg": true,
        ".png":  true,
        ".gif":  true,
        ".pdf":  true,
    }
    
    if !allowedExts[strings.ToLower(ext)] {
        return nil, zorya.Error400BadRequest("Invalid file extension")
    }
    
    // Process file...
}
```

## OpenAPI Documentation

Zorya automatically generates OpenAPI 3.1 documentation for file uploads:

```json
{
  "requestBody": {
    "content": {
      "multipart/form-data": {
        "schema": {
          "type": "object",
          "properties": {
            "File": {
              "type": "string",
              "format": "binary",
              "contentMediaType": "application/octet-stream",
              "description": "File content"
            },
            "Filename": {
              "type": "string",
              "description": "Name of the file"
            }
          }
        },
        "encoding": {
          "File": {
            "contentType": "application/octet-stream"
          }
        }
      }
    },
    "required": true
  }
}
```

## Streaming File Downloads

For file downloads, use streaming responses:

```go
type DownloadFileInput struct {
    FileID string `schema:"file_id,location=path,required=true"`
}

type DownloadFileOutput struct {
    Status      int    `status:"200"`
    ContentType string `header:"Content-Type"`
    ContentDisposition string `header:"Content-Disposition"`
    Body        func(ctx zorya.Context)
}

func downloadFile(ctx context.Context, input *DownloadFileInput) (*DownloadFileOutput, error) {
    // Get file metadata
    file, err := fileService.GetByID(ctx, input.FileID)
    if err != nil {
        return nil, zorya.Error404NotFound("File not found")
    }
    
    return &DownloadFileOutput{
        Status:      http.StatusOK,
        ContentType: file.ContentType,
        ContentDisposition: fmt.Sprintf(`attachment; filename="%s"`, file.Filename),
        Body: func(ctx zorya.Context) {
            // Stream file content
            w := ctx.BodyWriter()
            fileService.StreamTo(ctx.Context(), input.FileID, w)
        },
    }, nil
}
```

## Complete Example

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    "path/filepath"
    
    "github.com/go-chi/chi/v5"
    "github.com/talav/talav/pkg/component/zorya"
    "github.com/talav/talav/pkg/component/zorya/adapters"
)

type UploadFileInput struct {
    ResourceID string `schema:"resource_id,location=path,required=true"`
    Overwrite  bool   `schema:"overwrite,location=query"`
    
    Body struct {
        File        []byte `json:"file" openapi:"format=binary,description=File content"`
        Filename    string `json:"filename" validate:"required" openapi:"description=Name of the file"`
        Description string `json:"description" openapi:"description=Optional file description"`
    } `body:"multipart,required"`
}

type UploadFileOutput struct {
    Status int `status:"201"`
    Body struct {
        ID       int64  `json:"id"`
        Filename string `json:"filename"`
        Size     int64  `json:"size"`
        URL      string `json:"url"`
    } `body:"structured"`
}

func main() {
    router := chi.NewMux()
    adapter := adapters.NewChi(router)
    api := zorya.NewAPI(adapter)
    
    zorya.Post(api, "/resources/{resource_id}/files", uploadFile,
        func(route *zorya.BaseRoute) {
            route.MaxBodyBytes = 10 * 1024 * 1024 // 10MB
            route.BodyReadTimeout = 30 * time.Second
        },
    )
    
    http.ListenAndServe(":8080", router)
}

func uploadFile(ctx context.Context, input *UploadFileInput) (*UploadFileOutput, error) {
    // Validate file type
    contentType := http.DetectContentType(input.Body.File)
    allowedTypes := map[string]bool{
        "image/jpeg": true,
        "image/png":  true,
        "application/pdf": true,
    }
    
    if !allowedTypes[contentType] {
        return nil, zorya.Error400BadRequest("Invalid file type")
    }
    
    // Validate file extension
    ext := filepath.Ext(input.Body.Filename)
    allowedExts := map[string]bool{
        ".jpg": true, ".jpeg": true, ".png": true, ".pdf": true,
    }
    
    if !allowedExts[strings.ToLower(ext)] {
        return nil, zorya.Error400BadRequest("Invalid file extension")
    }
    
    // Save file (example - implement your own storage)
    fileID := saveFile(input.Body.File, input.Body.Filename, input.Body.Description)
    
    return &UploadFileOutput{
        Status: http.StatusCreated,
        Body: struct {
            ID       int64  `json:"id"`
            Filename string `json:"filename"`
            Size     int64  `json:"size"`
            URL      string `json:"url"`
        }{
            ID:       fileID,
            Filename: input.Body.Filename,
            Size:     int64(len(input.Body.File)),
            URL:      fmt.Sprintf("/files/%d", fileID),
        },
    }, nil
}
```

## Client Example

### cURL

```bash
curl -X POST http://localhost:8080/resources/123/files \
  -F "file=@/path/to/file.pdf" \
  -F "filename=document.pdf" \
  -F "description=Important document"
```

### Go Client

```go
package main

import (
    "bytes"
    "io"
    "mime/multipart"
    "net/http"
    "os"
)

func uploadFile(resourceID, filename string) error {
    // Open file
    file, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer file.Close()
    
    // Create multipart writer
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    
    // Add file field
    fileWriter, err := writer.CreateFormFile("file", filename)
    if err != nil {
        return err
    }
    _, err = io.Copy(fileWriter, file)
    if err != nil {
        return err
    }
    
    // Add other fields
    writer.WriteField("filename", filename)
    writer.WriteField("description", "Uploaded via Go client")
    
    err = writer.Close()
    if err != nil {
        return err
    }
    
    // Send request
    req, err := http.NewRequest("POST", 
        "http://localhost:8080/resources/"+resourceID+"/files", 
        body)
    if err != nil {
        return err
    }
    req.Header.Set("Content-Type", writer.FormDataContentType())
    
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    return nil
}
```

## See Also

- [Input Structs](input-structs.md) - Request handling fundamentals
- [Validation](validation.md) - Input validation
- [Limits](limits.md) - Body size and timeout configuration
- [Streaming](../responses/streaming.md) - Streaming responses for downloads
