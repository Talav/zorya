package zorya

// Config represents a configuration for a new API. See `DefaultConfig()`
// as a starting point.
type Config struct {
	// OpenAPIPath is the exact path to the OpenAPI spec endpoint.
	// The path is used as-is, with no automatic suffix or extension handling.
	OpenAPIPath string

	// DocsPath is the path to the API documentation. If set to `/docs` it will
	// allow clients to get `/docs` to view the documentation in a browser. If
	// you wish to provide your own documentation renderer, you can leave this
	// blank and attach it directly to the router or adapter.
	DocsPath string

	// SchemasPath is the path to the API schemas. If set to `/schemas` it will
	// allow clients to get `/schemas/{schema}` to view the schema in a browser
	// or for use in editors like VSCode to provide autocomplete & validation.
	SchemasPath string

	// DefaultFormat specifies the default content type to use when the client
	// does not specify one. If unset, the default type will be randomly
	// chosen from the keys of `Formats`.
	DefaultFormat string

	// NoFormatFallback disables the fallback to application/json (if available)
	// when the client requests an unknown content type. If set and no format is
	// negotiated, then a 406 Not Acceptable response will be returned.
	NoFormatFallback bool
}

// DefaultConfig returns a default configuration for a new API. It is a good
// starting point for creating your own configuration. It supports the JSON
// format out of the box. The registry uses references for structs and a link
// transformer is included to add `$schema` fields and links into responses. The
// `/openapi.json`, `/docs`, and `/schemas` paths are set up to serve the
// OpenAPI spec, docs UI, and schemas respectively.
//
//	// Create and customize the config (if desired).
//	config := zorya.DefaultConfig()
//
//	// Create the API using the config.
//	router := adapters.NewChiAdapter()
//	api := zorya.NewAPI(router, zorya.WithConfig(config))
func DefaultConfig() *Config {
	return &Config{
		OpenAPIPath:      "/openapi.json",
		DocsPath:         "/docs",
		SchemasPath:      "/schemas",
		DefaultFormat:    "application/json",
		NoFormatFallback: false,
	}
}
