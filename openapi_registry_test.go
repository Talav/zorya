package zorya

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStruct is a test struct with 5 fields for e2e testing.
type TestStruct struct {
	ID       int      `schema:"id,required" openapi:"title=User ID,description=Unique identifier"`
	Name     string   `schema:"name,required" validate:"min=1,max=100" default:"Unknown"`
	Email    string   `schema:"email" openapi:"title=Email Address,description=User email address" validate:"email"`
	Age      *int     `schema:"age" openapi:"title=Age,description=User age in years" validate:"min=0,max=150"`
	IsActive bool     `schema:"is_active" default:"true" openapi:"readOnly"`
	_        struct{} `openapiStruct:"additionalProperties=false"`
}

func TestRegistry_EndToEnd(t *testing.T) {
	// Create metadata with all parsers
	metadata := NewMetadata()

	// Create registry
	registry := NewMapRegistry("#/components/schemas/", DefaultSchemaNamer, metadata)

	// Generate schema for TestStruct
	testType := reflect.TypeOf(TestStruct{})
	_ = registry.Schema(testType, true, "TestStruct")

	// Expected JSON string - review and verify this is correct
	wantedJSON := `{
  "TestStruct": {
    "properties": {
      "age": {
        "description": "User age in years",
        "maximum": 150,
        "minimum": 0,
        "title": "Age",
        "type": [
          "integer",
          "null"
        ]
      },
      "email": {
        "description": "User email address",
        "format": "email",
        "title": "Email Address",
        "type": "string"
      },
      "id": {
        "description": "Unique identifier",
        "title": "User ID",
        "type": "integer"
      },
      "is_active": {
        "default": true,
        "readOnly": true,
        "type": "boolean"
      },
      "name": {
        "default": "Unknown",
        "maxLength": 100,
        "minLength": 1,
        "type": "string"
      }
    },
    "required": [
      "id",
      "name"
    ],
    "type": "object"
  }
}`

	// Marshal generated JSON with indentation for comparison
	registryJSONIndented, err := json.MarshalIndent(registry, "", "  ")
	require.NoError(t, err, "Should be able to marshal registry with indentation")

	// Compare JSON strings
	assert.JSONEq(t, wantedJSON, string(registryJSONIndented), "Generated JSON should match expected JSON")
}

// TestRegistry_NestedStructs tests schema references and nested struct handling.
func TestRegistry_NestedStructs(t *testing.T) {
	type Address struct {
		Street  string `schema:"street,required" openapi:"title=Street Address"`
		City    string `schema:"city,required" openapi:"title=City"`
		Country string `schema:"country" default:"USA"`
	}

	type User struct {
		ID      int     `schema:"id,required"`
		Name    string  `schema:"name,required"`
		Address Address `schema:"address,required" openapi:"title=User Address"`
	}

	metadata := NewMetadata()
	registry := NewMapRegistry("#/components/schemas/", DefaultSchemaNamer, metadata)

	userType := reflect.TypeOf(User{})
	_ = registry.Schema(userType, true, "User")

	// Expected JSON string - review and verify this is correct
	wantedJSON := `{
  "Address": {
    "properties": {
      "city": {
        "title": "City",
        "type": "string"
      },
      "country": {
        "default": "USA",
        "type": "string"
      },
      "street": {
        "title": "Street Address",
        "type": "string"
      }
    },
    "required": [
      "street",
      "city"
    ],
    "type": "object"
  },
  "User": {
    "properties": {
      "address": {
        "$ref": "#/components/schemas/Address",
        "title": "User Address"
      },
      "id": {
        "format": "int64",
        "type": "integer"
      },
      "name": {
        "type": "string"
      }
    },
    "required": [
      "id",
      "name",
      "address"
    ],
    "type": "object"
  }
}`

	// Marshal generated JSON with indentation for comparison
	registryJSONIndented, err := json.MarshalIndent(registry, "", "  ")
	require.NoError(t, err, "Should be able to marshal registry with indentation")

	// Compare JSON strings
	assert.JSONEq(t, wantedJSON, string(registryJSONIndented), "Generated JSON should match expected JSON")
}

// TestRegistry_ArraysWithComplexItems tests array schema generation with complex item types.
func TestRegistry_ArraysWithComplexItems(t *testing.T) {
	type Tag struct {
		Name  string `schema:"name,required" validate:"min=1,max=50"`
		Color string `schema:"color" validate:"oneof=red blue green"`
	}

	type Post struct {
		ID    int    `schema:"id,required"`
		Title string `schema:"title,required" validate:"min=5,max=200"`
		Tags  []Tag  `schema:"tags" validate:"min=1,max=10" openapi:"title=Post Tags,description=List of tags"`
	}

	metadata := NewMetadata()
	registry := NewMapRegistry("#/components/schemas/", DefaultSchemaNamer, metadata)

	postType := reflect.TypeOf(Post{})
	_ = registry.Schema(postType, true, "Post")

	// Expected JSON string - review and verify this is correct
	wantedJSON := `{
  "Post": {
    "properties": {
      "id": {
        "format": "int64",
        "type": "integer"
      },
      "tags": {
        "description": "List of tags",
        "items": {
          "$ref": "#/components/schemas/Tag"
        },
        "maxItems": 10,
        "minItems": 1,
        "title": "Post Tags",
        "type": [
          "array",
          "null"
        ]
      },
      "title": {
        "maxLength": 200,
        "minLength": 5,
        "type": "string"
      }
    },
    "required": [
      "id",
      "title"
    ],
    "type": "object"
  },
  "Tag": {
    "properties": {
      "color": {
        "enum": [
          "red",
          "blue",
          "green"
        ],
        "type": "string"
      },
      "name": {
        "maxLength": 50,
        "minLength": 1,
        "type": "string"
      }
    },
    "required": [
      "name"
    ],
    "type": "object"
  }
}`

	// Marshal generated JSON with indentation for comparison
	registryJSONIndented, err := json.MarshalIndent(registry, "", "  ")
	require.NoError(t, err, "Should be able to marshal registry with indentation")

	// Compare JSON strings
	assert.JSONEq(t, wantedJSON, string(registryJSONIndented), "Generated JSON should match expected JSON")
}

// TestRegistry_DependentRequired tests dependentRequired feature with multiple dependencies.
func TestRegistry_DependentRequired(t *testing.T) {
	type PaymentInput struct {
		PaymentMethod    string `schema:"payment_method" dependentRequired:"billing_address,cardholder_name"`
		BillingAddress   string `schema:"billing_address"`
		CardholderName   string `schema:"cardholder_name"`
		ShippingAddress  string `schema:"shipping_address"`
		UseBillingAsShip bool   `schema:"use_billing_as_ship" dependentRequired:"shipping_address"`
	}

	metadata := NewMetadata()
	registry := NewMapRegistry("#/components/schemas/", DefaultSchemaNamer, metadata)

	paymentType := reflect.TypeOf(PaymentInput{})
	_ = registry.Schema(paymentType, true, "PaymentInput")

	// Expected JSON string - review and verify this is correct
	wantedJSON := `{
  "PaymentInput": {
    "dependentRequired": {
      "payment_method": [
        "billing_address",
        "cardholder_name"
      ],
      "use_billing_as_ship": [
        "shipping_address"
      ]
    },
    "properties": {
      "billing_address": {
        "type": "string"
      },
      "cardholder_name": {
        "type": "string"
      },
      "payment_method": {
        "type": "string"
      },
      "shipping_address": {
        "type": "string"
      },
      "use_billing_as_ship": {
        "type": "boolean"
      }
    },
    "type": "object"
  }
}`

	// Marshal generated JSON with indentation for comparison
	registryJSONIndented, err := json.MarshalIndent(registry, "", "  ")
	require.NoError(t, err, "Should be able to marshal registry with indentation")

	// Compare JSON strings
	assert.JSONEq(t, wantedJSON, string(registryJSONIndented), "Generated JSON should match expected JSON")
}

// TestRegistry_RecursiveTypes tests that recursive/circular type definitions are handled correctly
// without causing infinite loops or stack overflows.
func TestRegistry_RecursiveTypes(t *testing.T) {
	type Node struct {
		Next *Node `schema:"next"`
	}

	type Tree struct {
		Root *Node `schema:"root,required"`
	}

	metadata := NewMetadata()
	registry := NewMapRegistry("#/components/schemas/", DefaultSchemaNamer, metadata)

	treeType := reflect.TypeOf(Tree{})
	_ = registry.Schema(treeType, true, "Tree")

	// Expected JSON string - recursive types should use $ref to break cycles
	wantedJSON := `{
  "Node": {
    "properties": {
      "next": {
        "$ref": "#/components/schemas/Node"
      }
    },
    "type": "object"
  },
  "Tree": {
    "properties": {
      "root": {
        "$ref": "#/components/schemas/Node"
      }
    },
    "required": [
      "root"
    ],
    "type": "object"
  }
}`

	// Marshal generated JSON with indentation for comparison
	registryJSONIndented, err := json.MarshalIndent(registry, "", "  ")
	require.NoError(t, err, "Should be able to marshal registry with indentation")

	// Compare JSON strings
	assert.JSONEq(t, wantedJSON, string(registryJSONIndented), "Generated JSON should match expected JSON")

	// Verify that Node schema exists and references itself correctly
	var result map[string]any
	err = json.Unmarshal(registryJSONIndented, &result)
	require.NoError(t, err)

	nodeSchema, ok := result["Node"].(map[string]any)
	require.True(t, ok, "Node schema should exist")

	props, ok := nodeSchema["properties"].(map[string]any)
	require.True(t, ok)

	nextProp, ok := props["next"].(map[string]any)
	require.True(t, ok, "next property should exist")

	// Verify that next property references Node (breaking the cycle)
	assert.Contains(t, nextProp, "$ref", "next property should have $ref to break recursion")
	assert.Equal(t, "#/components/schemas/Node", nextProp["$ref"], "next should reference Node schema")
}
