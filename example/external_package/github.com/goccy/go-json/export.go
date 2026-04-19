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
			"Marshaler":          reflect.TypeFor[q.Marshaler](),
			"MarshalerContext":   reflect.TypeFor[q.MarshalerContext](),
			"Unmarshaler":        reflect.TypeFor[q.Unmarshaler](),
			"UnmarshalerContext": reflect.TypeFor[q.UnmarshalerContext](),
		},
		NamedTypes: map[string]reflect.Type{
			"DecodeOptionFunc": reflect.TypeFor[q.DecodeOptionFunc](),
			"Decoder":          reflect.TypeFor[q.Decoder](),
			"EncodeOptionFunc": reflect.TypeFor[q.EncodeOptionFunc](),
			"Encoder":          reflect.TypeFor[q.Encoder](),
			"Path":             reflect.TypeFor[q.Path](),
			"SubFieldQuery":    reflect.TypeFor[q.SubFieldQuery](),
		},
		AliasTypes: map[string]reflect.Type{
			"ColorFormat":           reflect.TypeFor[q.ColorFormat](),
			"ColorScheme":           reflect.TypeFor[q.ColorScheme](),
			"DecodeOption":          reflect.TypeFor[q.DecodeOption](),
			"Delim":                 reflect.TypeFor[q.Delim](),
			"EncodeOption":          reflect.TypeFor[q.EncodeOption](),
			"FieldQuery":            reflect.TypeFor[q.FieldQuery](),
			"FieldQueryString":      reflect.TypeFor[q.FieldQueryString](),
			"InvalidUTF8Error":      reflect.TypeFor[q.InvalidUTF8Error](),
			"InvalidUnmarshalError": reflect.TypeFor[q.InvalidUnmarshalError](),
			"MarshalerError":        reflect.TypeFor[q.MarshalerError](),
			"Number":                reflect.TypeFor[q.Number](),
			"PathError":             reflect.TypeFor[q.PathError](),
			"RawMessage":            reflect.TypeFor[q.RawMessage](),
			"SyntaxError":           reflect.TypeFor[q.SyntaxError](),
			"Token":                 reflect.TypeFor[q.Token](),
			"UnmarshalFieldError":   reflect.TypeFor[q.UnmarshalFieldError](),
			"UnmarshalTypeError":    reflect.TypeFor[q.UnmarshalTypeError](),
			"UnsupportedTypeError":  reflect.TypeFor[q.UnsupportedTypeError](),
			"UnsupportedValueError": reflect.TypeFor[q.UnsupportedValueError](),
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
