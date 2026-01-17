package metadata

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseOpenAPIStructTag(t *testing.T) {
	tests := []struct {
		name        string
		fieldName   string
		tagValue    string
		want        *OpenAPIStructMetadata
		wantErr     bool
		errContains string
	}{
		{
			name:      "empty tag",
			fieldName: "_",
			tagValue:  "",
			want:      &OpenAPIStructMetadata{},
		},
		{
			name:      "additionalProperties true",
			fieldName: "_",
			tagValue:  "additionalProperties=true",
			want: &OpenAPIStructMetadata{
				AdditionalProperties: boolPtr(true),
			},
		},
		{
			name:      "additionalProperties false",
			fieldName: "_",
			tagValue:  "additionalProperties=false",
			want: &OpenAPIStructMetadata{
				AdditionalProperties: boolPtr(false),
			},
		},
		{
			name:      "additionalProperties flag (true)",
			fieldName: "_",
			tagValue:  "additionalProperties",
			want: &OpenAPIStructMetadata{
				AdditionalProperties: boolPtr(true),
			},
		},
		{
			name:      "nullable true",
			fieldName: "_",
			tagValue:  "nullable=true",
			want: &OpenAPIStructMetadata{
				Nullable: boolPtr(true),
			},
		},
		{
			name:      "nullable false",
			fieldName: "_",
			tagValue:  "nullable=false",
			want: &OpenAPIStructMetadata{
				Nullable: boolPtr(false),
			},
		},
		{
			name:      "nullable flag (true)",
			fieldName: "_",
			tagValue:  "nullable",
			want: &OpenAPIStructMetadata{
				Nullable: boolPtr(true),
			},
		},
		{
			name:      "both options",
			fieldName: "_",
			tagValue:  "additionalProperties=false,nullable=true",
			want: &OpenAPIStructMetadata{
				AdditionalProperties: boolPtr(false),
				Nullable:             boolPtr(true),
			},
		},
		{
			name:      "both options reversed",
			fieldName: "_",
			tagValue:  "nullable=true,additionalProperties=false",
			want: &OpenAPIStructMetadata{
				AdditionalProperties: boolPtr(false),
				Nullable:             boolPtr(true),
			},
		},
		{
			name:      "unknown option ignored",
			fieldName: "_",
			tagValue:  "additionalProperties=true,unknown=value",
			want: &OpenAPIStructMetadata{
				AdditionalProperties: boolPtr(true),
			},
		},
		{
			name:        "invalid tag parsing",
			fieldName:   "_",
			tagValue:    "additionalProperties=true,'unclosed quote",
			wantErr:     true,
			errContains: "failed to parse openapiStruct tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := reflect.StructField{
				Name: tt.fieldName,
			}

			result, err := ParseOpenAPIStructTag(field, 0, tt.tagValue)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}

				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			osm, ok := result.(*OpenAPIStructMetadata)
			require.True(t, ok, "result should be *OpenAPIStructMetadata")

			assert.Equal(t, tt.want.AdditionalProperties, osm.AdditionalProperties, "AdditionalProperties mismatch")
			assert.Equal(t, tt.want.Nullable, osm.Nullable, "Nullable mismatch")
		})
	}
}
