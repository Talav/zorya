# Zorya Documentation

This directory contains the comprehensive documentation for the Zorya HTTP API framework.

## Documentation Structure

The documentation follows a progressive disclosure pattern:

```
docs/
├── index.md                          # Landing page
├── mkdocs.yml                        # MkDocs configuration
├── introduction/
│   ├── overview.md                   # Architecture and concepts
│   ├── why-zorya.md                  # Comparison with alternatives
│   ├── installation.md               # Getting started
│   └── architecture.md               # Technical architecture
├── tutorial/
│   ├── quick-start.md                # 5-minute quickstart
│   ├── first-api.md                  # Building your first API
│   ├── validation.md                 # Adding validation
│   ├── security.md                   # Adding authentication
│   └── testing.md                    # Testing your API
├── features/
│   ├── features-overview.md          # Complete feature list
│   ├── router-adapters.md            # Chi, Fiber, Stdlib
│   ├── requests/
│   │   ├── input-structs.md          # Request handling
│   │   ├── validation.md             # Input validation
│   │   ├── file-uploads.md           # File upload support
│   │   └── limits.md                 # Body limits and timeouts
│   ├── responses/
│   │   ├── output-structs.md         # Response encoding
│   │   ├── errors.md                 # RFC 9457 error handling
│   │   ├── streaming.md              # SSE and chunked responses
│   │   └── transformers.md           # Response transformers
│   ├── content-negotiation.md        # JSON, CBOR, custom formats
│   ├── security/
│   │   ├── overview.md               # Security architecture
│   │   ├── authentication.md         # JWT auth
│   │   ├── authorization.md          # Roles and permissions
│   │   └── resource-based.md         # Resource-level RBAC
│   ├── middleware.md                 # Middleware patterns
│   ├── groups.md                     # Route groups
│   ├── conditional-requests.md       # ETags, If-Match
│   ├── defaults.md                   # Default parameter values
│   ├── openapi/
│   │   ├── overview.md               # OpenAPI generation
│   │   ├── documentation-ui.md       # Interactive docs UI
│   │   └── schema-generation.md      # Schema customization
│   └── metadata/
│       ├── overview.md               # Metadata system
│       └── tags-reference.md         # All struct tags
├── how-to/
│   ├── custom-validators.md          # Implement custom validation
│   ├── custom-formats.md             # Add XML, YAML, etc.
│   ├── custom-enforcers.md           # Custom authorization
│   ├── graceful-shutdown.md          # Production shutdown
│   ├── fx-integration.md             # Uber FX integration
│   └── testing.md                    # Testing strategies
├── reference/
│   ├── api.md                        # Complete API reference
│   ├── context.md                    # Context interface
│   ├── types.md                      # Type definitions
│   └── constants.md                  # Constants
└── packages/
    ├── schema.md                     # Schema package
    ├── negotiation.md                # Negotiation package
    ├── validator.md                  # Validator package
    ├── security.md                   # Security component
    └── conditional.md                # Conditional package
```

## Building the Documentation

### Install MkDocs

```bash
pip install mkdocs-material
```

### Serve Locally

```bash
cd docs
mkdocs serve
```

Visit `http://localhost:8000`

### Build Static Site

```bash
cd docs
mkdocs build
```

Output is in `site/` directory.

### Deploy to GitHub Pages

```bash
cd docs
mkdocs gh-deploy
```

## Documentation Philosophy

### Progressive Disclosure

1. **Introduction** - High-level overview, why Zorya exists
2. **Tutorial** - Step-by-step guide, learn by doing
3. **Features** - Deep dive into capabilities
4. **How-To** - Solutions to specific problems
5. **Reference** - Complete API documentation

### Code as Documentation

Documentation examples are extracted from actual tests. When code changes, documentation must be updated.

### User-Centric

- Start with what users want to achieve
- Show working code first, explain concepts after
- Link to related topics
- Provide complete, runnable examples

## Current Status

### ✅ Completed

- [x] Documentation structure
- [x] Landing page (index.md)
- [x] Introduction section (overview)
- [x] MkDocs configuration
- [x] File uploads documentation
- [x] Documentation UI documentation
- [x] Features overview with complete feature list

### 🚧 In Progress

- [ ] Complete tutorial section
- [ ] Complete all feature pages
- [ ] Complete how-to guides
- [ ] Complete reference documentation
- [ ] Add diagrams and visualizations
- [ ] Add more examples

### 📋 Todo

- [ ] Split existing README.md content into feature pages
- [ ] Create missing feature documentation
- [ ] Add tutorial walkthroughs
- [ ] Add how-to guides for common scenarios
- [ ] Create complete API reference
- [ ] Add troubleshooting guide
- [ ] Add migration guides
- [ ] Add performance tuning guide

## Missing Features in Current README

The following features exist in code but are missing from the old README:

1. ✅ **File Uploads** (multipart/form-data) - NOW DOCUMENTED
2. ✅ **Documentation UI** (Stoplight Elements) - NOW DOCUMENTED
3. ⏳ **OpenAPI endpoint** (/openapi.json, /openapi.yaml)
4. ⏳ **Multiple content types** (detailed CBOR support)
5. ⏳ **Encoding configuration** for multipart
6. ⏳ **Binary format support** (contentMediaType)
7. ⏳ **Dependent required** fields (JSON Schema)
8. ⏳ **OpenAPI struct-level metadata** (additionalProperties, nullable)
9. ⏳ **Security schemes** configuration
10. ⏳ **External documentation** links

## Contributing

When adding new features:

1. Add feature documentation to appropriate section
2. Update features-overview.md
3. Add working examples
4. Update navigation in mkdocs.yml
5. Add to main README.md if essential

## Related Documentation

- Main README: `../README.md` (kept minimal, links here)
- Schema Package: `../../schema/README.md`
- Security Component: `../../security/README.md`
- Validator Package: `../../validator/README.md`

## Long-Term Goal

Deploy to dedicated site: `https://zorya.rocks/`

Similar to:
- [FastAPI](https://fastapi.tiangolo.com/)
- [NestJS](https://nestjs.com/)
