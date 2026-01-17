package zorya

import (
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"unicode/utf8"

	"github.com/talav/schema"
)

// Registry creates and stores schemas and their references, and supports
// marshalling to JSON for use as an OpenAPI #/components/schemas object.
// Behavior is implementation-dependent, but the design allows for recursive
// schemas to exist while being flexible enough to support other use cases
// like only inline objects (no refs) or always using refs for structs.
type Registry interface {
	Schema(t reflect.Type, allowRef bool, hint string) *Schema
	SchemaFromRef(ref string) *Schema
	TypeFromRef(ref string) reflect.Type
	Map() map[string]*Schema
	RegisterTypeAlias(t reflect.Type, alias reflect.Type)
	MarkInlineOnly(t reflect.Type, hint string)
}

// SchemaNamer provides schema names for types.
type SchemaNamer interface {
	Name(t reflect.Type, hint string) string
}

// SchemaNamerFunc is a function adapter that implements SchemaNamer.
type SchemaNamerFunc func(t reflect.Type, hint string) string

type mapRegistry struct {
	prefix     string
	schemas    map[string]*Schema
	types      map[string]reflect.Type
	seen       map[reflect.Type]bool
	inlineOnly map[string]bool // Schemas that should not be exported to components
	namer      SchemaNamerFunc
	aliases    map[reflect.Type]reflect.Type
	metadata   *schema.Metadata
	builder    SchemaBuilder
}

// NewMapRegistry creates a new registry that stores schemas in a map and
// returns references to them using the given prefix.
func NewMapRegistry(prefix string, namer SchemaNamerFunc, m *schema.Metadata) Registry {
	reg := &mapRegistry{
		prefix:     prefix,
		schemas:    map[string]*Schema{},
		types:      map[string]reflect.Type{},
		seen:       map[reflect.Type]bool{},
		inlineOnly: map[string]bool{},
		aliases:    map[reflect.Type]reflect.Type{},
		namer:      namer,
		metadata:   m,
	}
	// Create a single schema builder instance for reuse
	reg.builder = newSchemaBuilder(reg, m)

	return reg
}

// DefaultSchemaNamer provides schema names for types. It uses the type name
// when possible, ignoring the package name. If the type is generic, e.g.
// `MyType[SubType]`, then the brackets are removed like `MyTypeSubType`.
// If the type is unnamed, then the name hint is used.
// Note: if you plan to use types with the same name from different packages,
// you should implement your own namer function to prevent issues. Nested
// anonymous types can also present naming issues.
func DefaultSchemaNamer(t reflect.Type, hint string) string {
	name := deref(t).Name()

	if name == "" {
		name = hint
	}

	// Better support for lists, so e.g. `[]int` becomes `ListInt`.
	name = strings.ReplaceAll(name, "[]", "List[")

	result := ""
	for _, part := range strings.FieldsFunc(name, func(r rune) bool {
		// Split on special characters. Note that `,` is used when there are
		// multiple inputs to a generic type.
		return r == '[' || r == ']' || r == '*' || r == ','
	}) {
		// Split fully qualified names like `github.com/foo/bar.Baz` into `Baz`.
		fqn := strings.Split(part, ".")
		base := fqn[len(fqn)-1]

		// Add to result, and uppercase for better scalar support (`int` -> `Int`).
		// Use unicode-aware uppercase to support non-ASCII characters.
		r, size := utf8.DecodeRuneInString(base)
		result += strings.ToUpper(string(r)) + base[size:]
	}
	name = result

	return name
}

//nolint:cyclop // Complex function handling type registration and schema generation - acceptable complexity
func (r *mapRegistry) Schema(t reflect.Type, allowRef bool, hint string) *Schema {
	origType := t
	t = deref(t)

	// Pointer to array should decay to array
	if t.Kind() == reflect.Array || t.Kind() == reflect.Slice {
		origType = t
	}

	alias, ok := r.aliases[t]
	if ok {
		return r.Schema(alias, allowRef, hint)
	}

	getsRef := t.Kind() == reflect.Struct
	if t == timeType {
		// Special case: time.Time is always a string.
		getsRef = false
	}

	v := reflect.New(t).Interface()
	if _, ok := v.(SchemaProvider); ok {
		// Special case: type provides its own schema
		getsRef = false
	}
	if _, ok := v.(encoding.TextUnmarshaler); ok {
		// Special case: type can be unmarshalled from text so will be a `string`
		// and doesn't need a ref. This simplifies the schema a little bit.
		getsRef = false
	}

	name := r.namer(origType, hint)

	//nolint:nestif // Complex nested logic for reference handling - acceptable complexity
	if getsRef {
		if s, ok := r.schemas[name]; ok {
			if _, ok := r.seen[t]; !ok {
				// Name matches but type is different, so we have a dupe.

				panic(fmt.Errorf("duplicate name: %s, new type: %s, existing type: %s", name, t, r.types[name]))
			}
			if allowRef {
				return &Schema{Ref: r.prefix + name}
			}

			return s
		}
	}

	// First, register the type so refs can be created above for recursive types.
	if getsRef {
		r.schemas[name] = &Schema{}
		r.types[name] = t
		r.seen[t] = true
	}
	s, err := r.builder.SchemaFromType(origType)
	if err != nil {
		panic(fmt.Errorf("failed to generate schema for type %s: %w", origType, err))
	}
	if getsRef {
		r.schemas[name] = s
	}

	if getsRef && allowRef {
		return &Schema{Ref: r.prefix + name}
	}

	return s
}

func (r *mapRegistry) SchemaFromRef(ref string) *Schema {
	if !strings.HasPrefix(ref, r.prefix) {
		return nil
	}

	return r.schemas[ref[len(r.prefix):]]
}

func (r *mapRegistry) TypeFromRef(ref string) reflect.Type {
	return r.types[ref[len(r.prefix):]]
}

func (r *mapRegistry) Map() map[string]*Schema {
	// Filter out inline-only schemas
	result := make(map[string]*Schema, len(r.schemas))
	for name, schema := range r.schemas {
		if !r.inlineOnly[name] {
			result[name] = schema
		}
	}

	return result
}

func (r *mapRegistry) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.schemas)
}

// RegisterTypeAlias(t, alias) makes the schema generator use the `alias` type instead of `t`.
func (r *mapRegistry) RegisterTypeAlias(t reflect.Type, alias reflect.Type) {
	r.aliases[t] = alias
}

// MarkInlineOnly marks a schema as inline-only, excluding it from components/schemas.
// This is useful for schemas that need transformation (e.g., multipart bodies)
// where the component schema would differ from the inline usage.
func (r *mapRegistry) MarkInlineOnly(t reflect.Type, hint string) {
	t = deref(t)
	name := r.namer(t, hint)
	r.inlineOnly[name] = true
}

func deref(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t
}
