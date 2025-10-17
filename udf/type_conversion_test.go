package udf

import (
	"database/sql/driver"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/duckdb/duckdb-go/v2"
)

func TestGoTypeToDuckDBTypeInfo(t *testing.T) {
	type SimpleStruct struct {
		I int32
		S string
	}
	type NestedStruct struct {
		Name     string
		Simple   SimpleStruct
		Optional *SimpleStruct // Pointer to struct, will be dereferenced by goTypeToDuckDBTypeInfo
	}
	type StructWithUnexported struct {
		Exported string
	}
	type EmptyStruct struct{}
	// For map key testing, ensure it's a comparable type for Go maps
	type MapKeyStruct struct {
		KeyPart string
	}

	tests := []struct {
		name               string
		goType             reflect.Type
		expectedDuckDBType duckdb.Type // For simple types or outer type of complex ones
		expectError        bool
		errorContains      string // Substring to check in error message
		// For structs, we can add more specific checks later if needed (e.g. field count/names)
	}{
		{"int", reflect.TypeOf(int(0)), duckdb.TYPE_INTEGER, false, ""},
		{"int64", reflect.TypeOf(int64(0)), duckdb.TYPE_BIGINT, false, ""},
		{"float32", reflect.TypeOf(float32(0)), duckdb.TYPE_FLOAT, false, ""},
		{"float64", reflect.TypeOf(float64(0)), duckdb.TYPE_DOUBLE, false, ""},
		{"string", reflect.TypeOf(""), duckdb.TYPE_VARCHAR, false, ""},
		{"bool", reflect.TypeOf(false), duckdb.TYPE_BOOLEAN, false, ""},
		{"[]byte", reflect.TypeOf([]byte{}), duckdb.TYPE_BLOB, false, ""},
		{"time.Time", reflect.TypeOf(time.Time{}), duckdb.TYPE_TIMESTAMP, false, ""},

		// Struct tests
		{"SimpleStruct", reflect.TypeOf(SimpleStruct{}), duckdb.TYPE_STRUCT, false, ""},
		{"NestedStruct", reflect.TypeOf(NestedStruct{}), duckdb.TYPE_STRUCT, false, ""},
		{"PointerToSimpleStruct", reflect.TypeOf(&SimpleStruct{}), duckdb.TYPE_STRUCT, false, ""}, // Test with pointer to struct
		{"StructWithUnexportedFields", reflect.TypeOf(StructWithUnexported{}), duckdb.TYPE_STRUCT, false, ""},
		{"EmptyStruct (error)", reflect.TypeOf(EmptyStruct{}), 0, true, "no exported fields"},

		// Map tests
		{"map[string]int32", reflect.TypeOf(map[string]int32{}), duckdb.TYPE_MAP, false, ""},
		{"map[int16]SimpleStruct", reflect.TypeOf(map[int16]SimpleStruct{}), duckdb.TYPE_MAP, false, ""},
		{"map[string]*SimpleStruct", reflect.TypeOf(map[string]*SimpleStruct{}), duckdb.TYPE_MAP, false, ""}, // Map with pointer to struct values
		// Test a map where the key is a struct (DuckDB might have restrictions on this, but our type conversion should attempt it)
		{"map[MapKeyStruct]string", reflect.TypeOf(map[MapKeyStruct]string{}), duckdb.TYPE_MAP, false, ""},
		{"map[string]map[int]string", reflect.TypeOf(map[string]map[int]string{}), duckdb.TYPE_MAP, false, ""}, // Nested map

		// Negative tests for basic types
		{"unsupported chan", reflect.TypeOf(make(chan int)), 0, true, "unsupported Go type kind for UDF: chan (specific type: chan int)"},
		{"unsupported func", reflect.TypeOf(func() {}), 0, true, "unsupported Go type kind for UDF: func (specific type: func())"},
		{"unsupported slice []int", reflect.TypeOf([]int{}), 0, true, "unsupported Go slice element type"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualGoType := tt.goType
			// For pointer types, goTypeToDuckDBTypeInfo is expected to dereference them once.
			// if (tt.name == "PointerToSimpleStruct" || strings.HasSuffix(tt.name, "*SimpleStruct")) && actualGoType.Kind() == reflect.Ptr {
			// This direct dereferencing in the test might be too simplistic if goTypeToDuckDBTypeInfo handles nested pointers.
			// However, for UDFs, we usually expect direct types or a single pointer to a struct.
			// The main goTypeToDuckDBTypeInfo handles reflect.Struct and reflect.Map, not pointers to them directly for creating TypeInfo.
			// It expects the Type of the element if the argument is a pointer.
			// This test setup might need adjustment if the Ptr handling in goTypeToDuckDBTypeInfo changes.
			// For now, the TypeInfo generator expects non-pointer types for struct/map definitions.
			// If a UDF argument is *MyStruct, funcType.In(i) gives *MyStruct, we need to pass MyStruct to goTypeToDuckDBTypeInfo.
			// This is implicitly handled by how BuildScalarUDF calls it if we had a BuildScalarUDF test for *Struct args.
			// Let's adjust goTypeToDuckDBTypeInfo to handle one level of pointer dereferencing for struct/map kinds.
			// This part of the test is tricky because the call path matters. Let's simplify the test for now and ensure goTypeToDuckDBTypeInfo is robust.
			// The current goTypeToDuckDBTypeInfo doesn't explicitly dereference. It expects the element type.
			// So, for a *SimpleStruct, this test should pass reflect.TypeOf(SimpleStruct{}) or handle it in goTypeToDuckDBTypeInfo.
			// I've added a case for `reflect.Ptr` in goTypeToDuckDBTypeInfo to handle this.
			// }

			typeInfo, err := goTypeToDuckDBTypeInfo(actualGoType)

			if (err != nil) != tt.expectError {
				t.Errorf("goTypeToDuckDBTypeInfo() for %s error = %v, expectError %v", tt.name, err, tt.expectError)
				return
			}
			if tt.expectError {
				if err == nil {
					t.Errorf("goTypeToDuckDBTypeInfo() for %s expected an error, but got nil", tt.name)
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("goTypeToDuckDBTypeInfo() for %s error '%v' does not contain expected substring '%s'", tt.name, err, tt.errorContains)
				}
				return // Done with error case
			}

			if typeInfo == nil {
				t.Errorf("goTypeToDuckDBTypeInfo() for %s returned nil TypeInfo for a supported type", tt.name)
				return
			}
			if typeInfo.InternalType() != tt.expectedDuckDBType {
				t.Errorf("goTypeToDuckDBTypeInfo() for %s expected DuckDB Type %v, got %v", tt.name, tt.expectedDuckDBType, typeInfo.InternalType())
			}
		})
	}
}

func TestConvertToReflectValue(t *testing.T) {
	type TestSimpleStruct struct {
		I int32
		S string
		B bool
	}
	type TestNestedStruct struct {
		Name   string
		Simple TestSimpleStruct
		Val    *int64 // Pointer field for nil testing
	}
	var int64PtrVal int64 = 12345

	tests := []struct {
		name          string
		sourceVal     driver.Value
		targetGoType  reflect.Type
		expectedVal   any // For non-error cases, the actual Go value expected
		expectError   bool
		errorContains string // Substring for error message validation
	}{
		// Basic types (mostly existing, ensure they are still fine)
		{"nil to interface", nil, reflect.TypeOf((*any)(nil)).Elem(), nil, false, ""},
		{"int64 to int", int64(123), reflect.TypeOf(int(0)), int(123), false, ""},
		{"float64 to float32", float64(123.45), reflect.TypeOf(float32(0)), float32(123.45), false, ""},
		{"string to string", "hello", reflect.TypeOf(""), "hello", false, ""},
		{"bool to bool", true, reflect.TypeOf(false), true, false, ""},
		{"[]byte to []byte", []byte("blob"), reflect.TypeOf([]byte{}), []byte("blob"), false, ""},
		{
			"int64 (micros) to time.Time",
			time.Date(2023, 1, 1, 12, 30, 0, 0, time.UTC).UnixMicro(),
			reflect.TypeOf(time.Time{}),
			time.Date(2023, 1, 1, 12, 30, 0, 0, time.UTC),
			false, "",
		},
		{"nil to int (error)", nil, reflect.TypeOf(int(0)), nil, true, "cannot convert nil (driver.Value) to Go type int"},
		{"string to int (error)", "not-an-int", reflect.TypeOf(int(0)), nil, true, "cannot convert DuckDB driver.Value"},
		{"int64 to string (error - disallowed implicit)", int64(65), reflect.TypeOf(""), nil, true, "cannot convert DuckDB driver.Value"},

		// Struct conversion tests
		{
			"map to SimpleStruct (success)",
			map[string]any{"I": int32(10), "S": "test", "B": true},
			reflect.TypeOf(TestSimpleStruct{}),
			TestSimpleStruct{I: 10, S: "test", B: true},
			false, "",
		},
		{
			"map to NestedStruct (success)",
			map[string]any{
				"Name":   "nested",
				"Simple": map[string]any{"I": int32(20), "S": "inner", "B": false},
				"Val":    &int64PtrVal,
			},
			reflect.TypeOf(TestNestedStruct{}),
			TestNestedStruct{Name: "nested", Simple: TestSimpleStruct{I: 20, S: "inner", B: false}, Val: &int64PtrVal},
			false, "",
		},
		{
			"map to SimpleStruct (missing field error)",
			map[string]any{"I": int32(10), "B": true}, // Missing S
			reflect.TypeOf(TestSimpleStruct{}),
			nil,
			true, "field 'S' missing",
		},
		{
			"map to SimpleStruct (type mismatch error for field)",
			map[string]any{"I": "not-an-int", "S": "test", "B": true},
			reflect.TypeOf(TestSimpleStruct{}),
			nil,
			true, "error converting field 'I'",
		},
		{
			"nil to SimpleStruct (expect zero struct)",
			nil,
			reflect.TypeOf(TestSimpleStruct{}),
			TestSimpleStruct{I: 0, S: "", B: false}, // Zero value for struct
			false, "",
		},
		{
			"wrong source type for struct (expect map[string]interface{})",
			int64(123),
			reflect.TypeOf(TestSimpleStruct{}),
			nil,
			true, "expected map[string]interface{}",
		},

		// Map conversion tests
		{
			"duckdb.Map to map[string]int32 (success)",
			duckdb.Map(map[any]any{"key1": int32(100), "key2": int32(200)}),
			reflect.TypeOf(map[string]int32{}),
			map[string]int32{"key1": 100, "key2": 200},
			false, "",
		},
		{
			"duckdb.Map to map[int]TestSimpleStruct (success)",
			duckdb.Map(map[any]any{
				int(1): map[string]any{"I": int32(1), "S": "one", "B": true},
			}),
			reflect.TypeOf(map[int]TestSimpleStruct{}),
			map[int]TestSimpleStruct{1: {I: 1, S: "one", B: true}},
			false, "",
		},
		{
			"duckdb.Map to map[string]int (key type mismatch error)",
			duckdb.Map(map[any]any{int32(1): "val"}), // Key is int32, target map key is string
			reflect.TypeOf(map[string]int{}),
			nil,
			true, "error converting map key",
		},
		{
			"duckdb.Map to map[string]int (value type mismatch error)",
			duckdb.Map(map[any]any{"key1": "not-an-int"}),
			reflect.TypeOf(map[string]int{}),
			nil,
			true, "error converting map value for key 'key1'",
		},
		{
			"nil to map[string]int (expect nil map)",
			nil,
			reflect.TypeOf(map[string]int{}),
			map[string]int(nil), // Corrected: Expected value is a typed nil map
			false, "",
		},
		{
			"wrong source type for map (expect duckdb.Map)",
			"not-a-duckdb-map",
			reflect.TypeOf(map[string]int{}),
			nil,
			true, "expected duckdb.Map from driver",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Running convertToReflectValue subtest: %s", tt.name)
			val, err := convertToReflectValue(tt.sourceVal, tt.targetGoType)
			if (err != nil) != tt.expectError {
				t.Errorf("convertToReflectValue() error = %v, expectError %v. Source: %#v", err, tt.expectError, tt.sourceVal)
				return
			}
			if tt.expectError {
				if err == nil {
					t.Errorf("convertToReflectValue() expected an error containing '%s', but got nil. Source: %#v", tt.errorContains, tt.sourceVal)
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("convertToReflectValue() error '%v' does not contain expected substring '%s'. Source: %#v", err, tt.errorContains, tt.sourceVal)
				}
				return // Done with error case
			}

			// For non-error cases, compare the actual value
			if val == (reflect.Value{}) { // Check if reflect.Value is zero, which shouldn't happen if no error
				t.Errorf("convertToReflectValue() returned zero reflect.Value without error. Source: %#v", tt.sourceVal)
				return
			}
			actual := val.Interface()
			if !reflect.DeepEqual(actual, tt.expectedVal) {
				t.Errorf("convertToReflectValue() got = %#v (type %T), want %#v (type %T). Source: %#v", actual, actual, tt.expectedVal, tt.expectedVal, tt.sourceVal)
			}
		})
	}
}
