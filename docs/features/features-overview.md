# Zorya Features

Complete list of Zorya framework features with links to detailed documentation.

## Core Features

### Request Handling
- **[Input Structs](requests/input-structs.md)** - Type-safe request parsing from path, query, headers, cookies, and body
- **[Validation](requests/validation.md)** - Automatic input validation using go-playground/validator with structured error responses
- **[File Uploads](requests/file-uploads.md)** - Multipart/form-data file uploads with binary content support
- **[Request Limits](requests/limits.md)** - Configurable body size limits and read timeouts to prevent abuse
- **[Default Values](defaults.md)** - Automatic application of default values from struct tags

### Response Handling
- **[Output Structs](responses/output-structs.md)** - Type-safe response encoding with status codes and headers
- **[Error Handling](responses/errors.md)** - RFC 9457 Problem Details with structured, machine-readable errors
- **[Streaming](responses/streaming.md)** - Server-Sent Events (SSE) and chunked transfer for real-time data
- **[Transformers](responses/transformers.md)** - Modify response bodies before serialization
- **[Content Negotiation](content-negotiation.md)** - Automatic format selection (JSON, CBOR, custom)

### Routing
- **[Router Adapters](router-adapters.md)** - Chi, Fiber, and Go 1.22+ stdlib support
- **[Route Groups](groups.md)** - Organize routes with shared prefixes, middleware, and configuration
- **[HTTP Methods](router-adapters.md#http-methods)** - GET, POST, PUT, PATCH, DELETE, HEAD support

### Security
- **[Overview](security/overview.md)** - Declarative security architecture
- **[Authentication](security/authentication.md)** - JWT token validation and user context
- **[Authorization](security/authorization.md)** - Role-based and permission-based access control
- **[Resource-Based Access](security/resource-based.md)** - Fine-grained resource-level authorization with dynamic templates

### Middleware
- **[API-Level Middleware](middleware.md#api-level)** - Global middleware for logging, metrics, CORS, etc.
- **[Route-Level Middleware](middleware.md#route-level)** - Per-route middleware for specific handling
- **[Group-Level Middleware](middleware.md#group-level)** - Middleware shared across route groups
- **[Security Middleware](security/overview.md#middleware)** - JWT authentication and authorization enforcement

### OpenAPI Documentation
- **[Overview](openapi/overview.md)** - OpenAPI 3.1 specification generation
- **[Documentation UI](openapi/documentation-ui.md)** - Interactive API explorer with Stoplight Elements
- **[Schema Generation](openapi/schema-generation.md)** - Automatic schema creation from Go types
- **[Metadata System](metadata/overview.md)** - Struct tag-based metadata extraction

### HTTP Standards Support
- **[Conditional Requests](conditional-requests.md)** - If-Match, If-None-Match, If-Modified-Since, If-Unmodified-Since
- **[Content Negotiation](content-negotiation.md)** - RFC 7231 compliant content type negotiation
- **[RFC 9457 Errors](responses/errors.md)** - Problem Details for HTTP APIs
- **[Standard Status Codes](responses/errors.md#status-codes)** - Proper use of HTTP status codes

## Advanced Features

### Type System
- **[Type-Safe Handlers](requests/input-structs.md#type-safety)** - Compile-time type checking
- **[Generic Functions](router-adapters.md#registration)** - Generic route registration
- **[Struct Tags](metadata/tags-reference.md)** - Comprehensive tag system for metadata

### Content Types
- **[JSON](content-negotiation.md#json)** - Built-in JSON encoding/decoding
- **[CBOR](content-negotiation.md#cbor)** - Built-in CBOR encoding/decoding
- **[Custom Formats](../how-to/custom-formats.md)** - Add XML, YAML, Protocol Buffers, etc.
- **[Multipart Form Data](requests/file-uploads.md)** - File uploads and form submission

### Validation
- **[Built-in Validators](requests/validation.md#built-in)** - Required, email, min/max, patterns
- **[Custom Validators](../how-to/custom-validators.md)** - Implement custom validation logic
- **[Structured Errors](requests/validation.md#error-response)** - Detailed validation error responses
- **[OpenAPI Integration](requests/validation.md#openapi)** - Validation rules in OpenAPI schemas

### Performance
- **[Metadata Caching](metadata/overview.md#caching)** - Type metadata cached for efficient request handling
- **[Zero Allocation Paths](../introduction/architecture.md#performance)** - Minimal allocations in hot paths
- **[Streaming Support](responses/streaming.md)** - Efficient large response handling
- **[Concurrent Request Handling](../introduction/architecture.md)** - Go's concurrency model

### Developer Experience
- **[Type Safety](../introduction/overview.md#type-safety)** - Catch errors at compile time
- **[Auto Documentation](openapi/overview.md)** - Documentation stays in sync with code
- **[Clear Error Messages](responses/errors.md)** - Helpful validation and runtime errors
- **[IDE Support](metadata/tags-reference.md)** - Struct tags work with Go IDE tools

### Testing
- **[Handler Testing](../how-to/testing.md#handler)** - Test handlers in isolation
- **[Integration Testing](../how-to/testing.md#integration)** - Test full request/response cycle
- **[Mock Validators](../how-to/testing.md#mocks)** - Mock validation for testing
- **[Mock Enforcers](../how-to/testing.md#mocks)** - Mock security for testing

### Extensibility
- **[Custom Adapters](router-adapters.md#custom)** - Implement adapters for other routers
- **[Custom Validators](../how-to/custom-validators.md)** - Use any validation library
- **[Custom Enforcers](../how-to/custom-enforcers.md)** - Implement custom authorization
- **[Custom Formats](../how-to/custom-formats.md)** - Add support for any serialization format
- **[Response Transformers](responses/transformers.md)** - Transform responses globally or per-route

### Integration
- **[Uber FX](../how-to/fx-integration.md)** - Dependency injection integration
- **[Graceful Shutdown](../how-to/graceful-shutdown.md)** - Clean server shutdown
- **[Context Propagation](../reference/context.md)** - Context through middleware and handlers
- **[Standard Library](router-adapters.md#stdlib)** - Works with Go standard library

## Feature Comparison Matrix

| Feature | Zorya | Gin | Echo | Fiber |
|---------|-------|-----|------|-------|
| Type-safe handlers | ✅ | ❌ | ❌ | ❌ |
| Auto OpenAPI generation | ✅ | ❌ | ❌ | ❌ |
| Built-in validation | ✅ | ✅ | ✅ | ✅ |
| Content negotiation | ✅ | ❌ | ❌ | ❌ |
| File upload support | ✅ | ✅ | ✅ | ✅ |
| Streaming responses | ✅ | ✅ | ✅ | ✅ |
| RFC 9457 errors | ✅ | ❌ | ❌ | ❌ |
| Conditional requests | ✅ | ❌ | ❌ | ❌ |
| Resource-based authz | ✅ | ❌ | ❌ | ❌ |
| Multiple routers | ✅ | ❌ | ❌ | ❌ |
| Default values | ✅ | ❌ | ❌ | ❌ |
| Response transformers | ✅ | ❌ | ❌ | ❌ |
| Interactive docs UI | ✅ | ❌ | ❌ | ❌ |

## Feature Categories

### Essential Features (Start Here)
1. [Router Adapters](router-adapters.md) - Choose your router
2. [Input Structs](requests/input-structs.md) - Parse requests
3. [Output Structs](responses/output-structs.md) - Send responses
4. [Validation](requests/validation.md) - Validate input
5. [Error Handling](responses/errors.md) - Handle errors

### Production Features
1. [Request Limits](requests/limits.md) - Prevent abuse
2. [Security](security/overview.md) - Authentication & authorization
3. [Middleware](middleware.md) - Logging, metrics, CORS
4. [Graceful Shutdown](../how-to/graceful-shutdown.md) - Clean shutdown
5. [Testing](../how-to/testing.md) - Test your API

### Advanced Features
1. [File Uploads](requests/file-uploads.md) - Handle files
2. [Streaming](responses/streaming.md) - Real-time data
3. [Custom Formats](../how-to/custom-formats.md) - Add formats
4. [Resource-Based Access](security/resource-based.md) - Fine-grained authz
5. [Transformers](responses/transformers.md) - Modify responses

### Documentation Features
1. [OpenAPI Generation](openapi/overview.md) - Auto-generate specs
2. [Documentation UI](openapi/documentation-ui.md) - Interactive explorer
3. [Schema Customization](openapi/schema-generation.md) - Customize schemas
4. [Metadata Tags](metadata/tags-reference.md) - All struct tags

## Roadmap

### Planned Features
- **GraphQL Support** - GraphQL endpoint generation
- **WebSocket Support** - WebSocket connection handling
- **Rate Limiting** - Built-in rate limiting middleware
- **API Versioning** - URL and header-based versioning
- **Request Tracing** - OpenTelemetry integration
- **Mock Server** - Generate mock server from OpenAPI
- **Client Generation** - Generate Go clients from OpenAPI
- **Additional Formats** - XML, Protocol Buffers, MessagePack

### Under Consideration
- **Auto PATCH** - Automatic PATCH endpoint generation
- **Bulk Operations** - Batch request/response handling
- **Pagination** - Standard pagination patterns
- **Search/Filter DSL** - Query DSL for filtering
- **Webhooks** - Webhook registration and delivery
- **Background Jobs** - Async job processing

## Get Help

- **[Tutorial](../tutorial/quick-start.md)** - Step-by-step guide
- **[How-To Guides](../how-to/custom-validators.md)** - Solutions to specific problems
- **[API Reference](../reference/api.md)** - Complete API documentation
- **[GitHub Issues](https://github.com/talav/talav/issues)** - Report bugs or request features
