package metadata

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDependentRequiredTag(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		tagValue  string
		want      *DependentRequiredMetadata
		wantErr   bool
	}{
		{
			name:      "single dependent",
			fieldName: "CreditCard",
			tagValue:  "billing_address",
			want: &DependentRequiredMetadata{
				Dependents: []string{"billing_address"},
			},
		},
		{
			name:      "multiple dependents",
			fieldName: "CreditCard",
			tagValue:  "billing_address,cardholder_name",
			want: &DependentRequiredMetadata{
				Dependents: []string{"billing_address", "cardholder_name"},
			},
		},
		{
			name:      "multiple dependents with spaces",
			fieldName: "CreditCard",
			tagValue:  "billing_address, cardholder_name, phone_number",
			want: &DependentRequiredMetadata{
				Dependents: []string{"billing_address", "cardholder_name", "phone_number"},
			},
		},
		{
			name:      "empty tag",
			fieldName: "Field",
			tagValue:  "",
			want: &DependentRequiredMetadata{
				Dependents: []string{},
			},
		},
		{
			name:      "whitespace only",
			fieldName: "Field",
			tagValue:  "   ",
			wantErr:   true, // tagparser doesn't allow whitespace-only keys
		},
		{
			name:      "quoted field names",
			fieldName: "Field",
			tagValue:  "'field1','field2 with spaces','field3'",
			want: &DependentRequiredMetadata{
				Dependents: []string{"field1", "field2 with spaces", "field3"},
			},
		},
		{
			name:      "quoted field with comma",
			fieldName: "Field",
			tagValue:  "field1,'field,with,comma',field2",
			want: &DependentRequiredMetadata{
				Dependents: []string{"field1", "field,with,comma", "field2"},
			},
		},
		{
			name:      "empty values filtered",
			fieldName: "Field",
			tagValue:  "field1,,field2,  ,field3",
			wantErr:   true, // tagparser doesn't allow empty keys
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := reflect.StructField{
				Name: tt.fieldName,
			}

			result, err := ParseDependentRequiredTag(field, 0, tt.tagValue)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			dm, ok := result.(*DependentRequiredMetadata)
			require.True(t, ok, "result should be *DependentRequiredMetadata")

			// Compare as sets (order may vary due to map iteration)
			assert.ElementsMatch(t, tt.want.Dependents, dm.Dependents, "Dependents mismatch")
		})
	}
}
