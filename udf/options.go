package udf

import (
	"fmt"
	"reflect"

	"github.com/marcboeker/go-duckdb/v2"
)

// udfOption contains optional parameters for BuildScalarUDF.
type udfOption struct {
	volatile            bool
	specialNullHandling bool
}

// WithVolatile sets whether the UDF is a volatile function.
// Volatile functions may return different results for the same inputs (e.g., random()).
// If a function always returns the same result for the same inputs, this option should not be set or set to false.
func WithVolatile(v bool) func(*udfOption) {
	return func(o *udfOption) {
		o.volatile = v
	}
}

// WithSpecialNullHandling sets whether to enable special NULL value handling logic.
// If true, the user-defined function (fn) will still be called when any input parameter is SQL NULL.
// If false (default behavior), when any input parameter is SQL NULL, DuckDB will propagate NULL directly and the user function fn will not be called.
func WithSpecialNullHandling(s bool) func(*udfOption) {
	return func(o *udfOption) {
		o.specialNullHandling = s
	}
}

// BuildScalarUDF builds a DuckDB scalar user-defined function (UDF) from a Go function.
//
// The fn parameter must be a Go function that meets the following requirements:
//   - Must return exactly one value
//   - Parameter types and return type must be DuckDB supported types, including:
//   - Basic types: including various Go integer types (such as int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64), float32, float64, string, bool, []byte. Integer types are automatically mapped to DuckDB's INTEGER or BIGINT based on their size and sign.
//   - time.Time
//   - structs (must have exported fields)
//   - map[K]V (K and V must be supported types)
//   - Can be a variadic function (e.g., func(fixed string, nums ...int))
//   - Pointers to structs, maps, or time.Time (e.g., *MyStruct, *map[string]int, *time.Time), which will be automatically dereferenced during type mapping.
//
// Options such as WithVolatile(true) or WithSpecialNullHandling(true) can be passed through the opts parameter to configure UDF behavior.
// By default, UDFs are non-volatile and do not use special NULL handling.
//
// Returns a UDF that implements the duckdb.ScalarFunc interface, which can be registered to DuckDB via RegisterScalarUDF.
//
// Returns an error if fn is not a function, returns multiple values, or uses unsupported types.
func BuildScalarUDF(fn any, opts ...func(*udfOption)) (duckdb.ScalarFunc, error) {
	options := &udfOption{
		volatile:            false, // Default value
		specialNullHandling: false, // Default value
	}
	for _, opt := range opts {
		opt(options)
	}

	funcVal := reflect.ValueOf(fn)
	funcType := funcVal.Type()

	if funcType.Kind() != reflect.Func {
		return nil, fmt.Errorf("BuildScalarUDF: input 'function' (type %s) is not a function, but %s", funcType.String(), funcType.Kind())
	}
	if funcType.NumOut() != 1 {
		return nil, fmt.Errorf("BuildScalarUDF: function (type %s) must return exactly one value, but returns %d", funcType.String(), funcType.NumOut())
	}

	numTotalGoArgs := funcType.NumIn()
	goArgTypes := make([]reflect.Type, numTotalGoArgs)
	for i := 0; i < numTotalGoArgs; i++ {
		goArgTypes[i] = funcType.In(i)
	}

	isGoFuncVariadic := funcType.IsVariadic()
	var duckDBInputTypeInfos []duckdb.TypeInfo     // For fixed args
	var duckDBVariadicElemTypeInfo duckdb.TypeInfo // For variadic arg element type

	numFixedArgs := numTotalGoArgs
	if isGoFuncVariadic {
		if numTotalGoArgs == 0 { // Should not happen if IsVariadic is true, but defensive
			return nil, fmt.Errorf("BuildScalarUDF: variadic function (type %s) has no arguments", funcType.String())
		}
		numFixedArgs--                                // Last Go arg is the variadic slice
		variadicSliceType := goArgTypes[numFixedArgs] // e.g. []int
		variadicElemType := variadicSliceType.Elem()  // e.g. int
		var err error
		duckDBVariadicElemTypeInfo, err = goTypeToDuckDBTypeInfo(variadicElemType)
		if err != nil {
			return nil, fmt.Errorf("BuildScalarUDF: error converting Go variadic element type for UDF (Go type %s, func type %s): %w", variadicElemType.String(), funcType.String(), err)
		}
	}

	duckDBInputTypeInfos = make([]duckdb.TypeInfo, numFixedArgs)
	for i := 0; i < numFixedArgs; i++ {
		goArgType := goArgTypes[i]
		duckDBTypeInfo, err := goTypeToDuckDBTypeInfo(goArgType)
		if err != nil {
			return nil, fmt.Errorf("BuildScalarUDF: error converting Go type for fixed argument %d of UDF (Go type %s, func type %s): %w", i, goArgType.String(), funcType.String(), err)
		}
		duckDBInputTypeInfos[i] = duckDBTypeInfo
	}

	goReturnType := funcType.Out(0)
	duckDBResultTypeInfo, err := goTypeToDuckDBTypeInfo(goReturnType)
	if err != nil {
		return nil, fmt.Errorf("BuildScalarUDF: error converting Go return type for UDF (Go type %s, func type %s): %w", goReturnType.String(), funcType.String(), err)
	}

	return &autoScalarFunc{
		userFunc:               funcVal,
		goArgTypes:             goArgTypes, // Store all Go arg types, including variadic slice type
		goReturnType:           goReturnType,
		duckDBInputTypeInfos:   duckDBInputTypeInfos, // Only fixed args for DuckDB config
		duckDBResultTypeInfo:   duckDBResultTypeInfo,
		specialNullHandling:    options.specialNullHandling,
		volatile:               options.volatile,
		isVariadic:             isGoFuncVariadic,
		duckDBVariadicTypeInfo: duckDBVariadicElemTypeInfo, // TypeInfo for the *element* of variadic part
	}, nil
}
