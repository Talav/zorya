package zorya

import (
	"encoding/json"
	"io"

	"github.com/fxamacker/cbor/v2"
)

// Format defines a marshal function for a content type.
type Format struct {
	Marshal func(w io.Writer, v any) error
}

// JSONFormat returns a Format for application/json.
func JSONFormat() Format {
	return Format{
		Marshal: func(w io.Writer, v any) error {
			enc := json.NewEncoder(w)
			enc.SetEscapeHTML(false)

			return enc.Encode(v)
		},
	}
}

// CBORFormat returns a Format for application/cbor.
func CBORFormat() Format {
	encMode, _ := cbor.EncOptions{
		Time: cbor.TimeRFC3339,
	}.EncMode()

	return Format{
		Marshal: func(w io.Writer, v any) error {
			return encMode.NewEncoder(w).Encode(v)
		},
	}
}

// DefaultFormats returns the standard format set with JSON and CBOR.
func DefaultFormats() map[string]Format {
	jsonFmt := JSONFormat()
	cborFmt := CBORFormat()

	return map[string]Format{
		"application/json": jsonFmt,
		"json":             jsonFmt, // For +json suffix matching
		"application/cbor": cborFmt,
		"cbor":             cborFmt, // For +cbor suffix matching
	}
}
