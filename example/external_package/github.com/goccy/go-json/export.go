// export by github.com/goplus/ixgo/cmd/qexp

package json

import (
	"reflect"

	q "github.com/goccy/go-json"

	"github.com/goplus/ixgo"
)

func init() {
	ixgo.RegisterPackage(&ixgo.Package{
		Name: "json",
		Path: "github.com/goccy/go-json",
		Deps: map[string]string{
			"bytes":         "bytes",
			"context":       "context",
			"encoding/json": "json",
			"fmt":           "fmt",
			"github.com/goccy/go-json/internal/decoder":                 "decoder",
			"github.com/goccy/go-json/internal/encoder":                 "encoder",
			"github.com/goccy/go-json/internal/encoder/vm":              "vm",
			"github.com/goccy/go-json/internal/encoder/vm_color":        "vm_color",
			"github.com/goccy/go-json/internal/encoder/vm_color_indent": "vm_color_indent",
			"github.com/goccy/go-json/internal/encoder/vm_indent":       "vm_indent",
			"github.com/goccy/go-json/internal/errors":                  "errors",
			"github.com/goccy/go-json/internal/runtime":                 "runtime",
			"io":      "io",
			"os":      "os",
			"reflect": "reflect",
			"unsafe":  "unsafe",
		},
		Interfaces: map[string]reflect.Type{
			"Marshaler":          reflect.TypeOf((*q.Marshaler)(nil)).Elem(),
			"MarshalerContext":   reflect.TypeOf((*q.MarshalerContext)(nil)).Elem(),
			"Unmarshaler":        reflect.TypeOf((*q.Unmarshaler)(nil)).Elem(),
			"UnmarshalerContext": reflect.TypeOf((*q.UnmarshalerContext)(nil)).Elem(),
		},
		NamedTypes: map[string]reflect.Type{
			"DecodeOptionFunc": reflect.TypeOf((*q.DecodeOptionFunc)(nil)).Elem(),
			"Decoder":          reflect.TypeOf((*q.Decoder)(nil)).Elem(),
			"EncodeOptionFunc": reflect.TypeOf((*q.EncodeOptionFunc)(nil)).Elem(),
			"Encoder":          reflect.TypeOf((*q.Encoder)(nil)).Elem(),
			"Path":             reflect.TypeOf((*q.Path)(nil)).Elem(),
			"SubFieldQuery":    reflect.TypeOf((*q.SubFieldQuery)(nil)).Elem(),
		},
		AliasTypes: map[string]reflect.Type{
			"ColorFormat":           reflect.TypeOf((*q.ColorFormat)(nil)).Elem(),
			"ColorScheme":           reflect.TypeOf((*q.ColorScheme)(nil)).Elem(),
			"DecodeOption":          reflect.TypeOf((*q.DecodeOption)(nil)).Elem(),
			"Delim":                 reflect.TypeOf((*q.Delim)(nil)).Elem(),
			"EncodeOption":          reflect.TypeOf((*q.EncodeOption)(nil)).Elem(),
			"FieldQuery":            reflect.TypeOf((*q.FieldQuery)(nil)).Elem(),
			"FieldQueryString":      reflect.TypeOf((*q.FieldQueryString)(nil)).Elem(),
			"InvalidUTF8Error":      reflect.TypeOf((*q.InvalidUTF8Error)(nil)).Elem(),
			"InvalidUnmarshalError": reflect.TypeOf((*q.InvalidUnmarshalError)(nil)).Elem(),
			"MarshalerError":        reflect.TypeOf((*q.MarshalerError)(nil)).Elem(),
			"Number":                reflect.TypeOf((*q.Number)(nil)).Elem(),
			"PathError":             reflect.TypeOf((*q.PathError)(nil)).Elem(),
			"RawMessage":            reflect.TypeOf((*q.RawMessage)(nil)).Elem(),
			"SyntaxError":           reflect.TypeOf((*q.SyntaxError)(nil)).Elem(),
			"Token":                 reflect.TypeOf((*q.Token)(nil)).Elem(),
			"UnmarshalFieldError":   reflect.TypeOf((*q.UnmarshalFieldError)(nil)).Elem(),
			"UnmarshalTypeError":    reflect.TypeOf((*q.UnmarshalTypeError)(nil)).Elem(),
			"UnsupportedTypeError":  reflect.TypeOf((*q.UnsupportedTypeError)(nil)).Elem(),
			"UnsupportedValueError": reflect.TypeOf((*q.UnsupportedValueError)(nil)).Elem(),
		},
		Vars: map[string]reflect.Value{
			"DefaultColorScheme":     reflect.ValueOf(&q.DefaultColorScheme),
			"FieldQueryFromContext":  reflect.ValueOf(&q.FieldQueryFromContext),
			"SetFieldQueryToContext": reflect.ValueOf(&q.SetFieldQueryToContext),
		},
		Funcs: map[string]reflect.Value{
			"BuildFieldQuery":             reflect.ValueOf(q.BuildFieldQuery),
			"BuildSubFieldQuery":          reflect.ValueOf(q.BuildSubFieldQuery),
			"Colorize":                    reflect.ValueOf(q.Colorize),
			"Compact":                     reflect.ValueOf(q.Compact),
			"CreatePath":                  reflect.ValueOf(q.CreatePath),
			"Debug":                       reflect.ValueOf(q.Debug),
			"DebugDOT":                    reflect.ValueOf(q.DebugDOT),
			"DebugWith":                   reflect.ValueOf(q.DebugWith),
			"DecodeFieldPriorityFirstWin": reflect.ValueOf(q.DecodeFieldPriorityFirstWin),
			"DisableHTMLEscape":           reflect.ValueOf(q.DisableHTMLEscape),
			"DisableNormalizeUTF8":        reflect.ValueOf(q.DisableNormalizeUTF8),
			"HTMLEscape":                  reflect.ValueOf(q.HTMLEscape),
			"Indent":                      reflect.ValueOf(q.Indent),
			"Marshal":                     reflect.ValueOf(q.Marshal),
			"MarshalContext":              reflect.ValueOf(q.MarshalContext),
			"MarshalIndent":               reflect.ValueOf(q.MarshalIndent),
			"MarshalIndentWithOption":     reflect.ValueOf(q.MarshalIndentWithOption),
			"MarshalNoEscape":             reflect.ValueOf(q.MarshalNoEscape),
			"MarshalWithOption":           reflect.ValueOf(q.MarshalWithOption),
			"NewDecoder":                  reflect.ValueOf(q.NewDecoder),
			"NewEncoder":                  reflect.ValueOf(q.NewEncoder),
			"Unmarshal":                   reflect.ValueOf(q.Unmarshal),
			"UnmarshalContext":            reflect.ValueOf(q.UnmarshalContext),
			"UnmarshalNoEscape":           reflect.ValueOf(q.UnmarshalNoEscape),
			"UnmarshalWithOption":         reflect.ValueOf(q.UnmarshalWithOption),
			"UnorderedMap":                reflect.ValueOf(q.UnorderedMap),
			"Valid":                       reflect.ValueOf(q.Valid),
		},
		TypedConsts:   map[string]ixgo.TypedConst{},
		UntypedConsts: map[string]ixgo.UntypedConst{},
	})
}
