package zorya

import (
	"fmt"
	"html"
	"net/http"
)

// registerDocsEndpoint registers the docs endpoint if configured.
func registerDocsEndpoint(a *api) {
	if a.config.DocsPath == "" {
		return
	}

	title := "API Documentation"
	if a.openAPI != nil && a.openAPI.Info != nil && a.openAPI.Info.Title != "" {
		title = a.openAPI.Info.Title
	}

	openAPIPath := a.config.OpenAPIPath
	if openAPIPath == "" {
		openAPIPath = "/openapi"
	}

	htmlContent := generateDocsHTML(openAPIPath+".json", title)

	a.adapter.Handle(&BaseRoute{
		Method: http.MethodGet,
		Path:   a.config.DocsPath,
	}, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(htmlContent))
	})
}

// generateDocsHTML generates an HTML page with embedded Stoplight Elements.
func generateDocsHTML(openAPIPath string, title string) string {
	escapedTitle := html.EscapeString(title)
	escapedPath := html.EscapeString(openAPIPath)

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>%s</title>
	<link rel="stylesheet" href="https://unpkg.com/@stoplight/elements/styles.min.css">
</head>
<body>
	<elements-api
		apiDescriptionUrl="%s"
		router="hash"
		layout="sidebar"
	/>
	<script src="https://unpkg.com/@stoplight/elements/web-components.min.js"></script>
</body>
</html>`, escapedTitle, escapedPath)
}
