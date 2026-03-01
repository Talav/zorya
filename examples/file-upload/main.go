// file-upload demonstrates multipart file uploads using *multipart.FileHeader.
// It validates file size and content type before accepting the upload.
//
// Run:
//
//	go run ./examples/file-upload
//
// Try:
//
//	# Upload an image (succeeds)
//	curl -X POST http://localhost:8080/uploads \
//	     -F "file=@/path/to/image.jpg"
//
//	# Upload with description
//	curl -X POST http://localhost:8080/uploads \
//	     -F "file=@/path/to/image.png" \
//	     -F "description=My profile photo"
//
//	# Too large (create a >2MB file to test)
//	dd if=/dev/urandom of=/tmp/big.bin bs=1M count=3
//	curl -X POST http://localhost:8080/uploads -F "file=@/tmp/big.bin"
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/talav/zorya"
	"github.com/talav/zorya/adapters"
)

const maxFileSize = 2 * 1024 * 1024 // 2 MB

// --- Input / output types ---

type UploadInput struct {
	Body struct {
		Description string                `schema:"description"`
		File        *multipart.FileHeader `schema:"file"`
	} `body:"multipart"`
}

type UploadOutput struct {
	Status int `json:"-"`
	Body   struct {
		Filename    string `json:"filename"`
		Size        int64  `json:"size"`
		ContentType string `json:"content_type"`
		Description string `json:"description,omitempty"`
	} `body:"structured"`
}

// --- Handler ---

func uploadFile(_ context.Context, input *UploadInput) (*UploadOutput, error) {
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

// --- Main ---

func main() {
	router := chi.NewMux()
	api := zorya.NewAPI(adapters.NewChi(router), zorya.WithConfig(zorya.DefaultConfig()))

	zorya.Post(api, "/uploads", uploadFile, func(r *zorya.BaseRoute) {
		r.MaxBodyBytes = 10 * 1024 * 1024 // 10 MB body limit (larger than per-file limit to allow metadata)
	})

	log.Println("Listening on :8080  —  docs at http://localhost:8080/docs")
	log.Fatal(http.ListenAndServe(":8080", router))
}
