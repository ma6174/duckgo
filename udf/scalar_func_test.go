package udf

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/duckdb/duckdb-go/v2"
)

func TestSuccessfulUDFRegistrationAndExecution(t *testing.T) {
	// Structs for UDF testing
	type MyStruct struct {
		ID   int32
		Name string
		Val  float64
	}
	type MyNestedStruct struct {
		Tag    string
		Nested MyStruct
	}

	// Define Go functions for UDFs
	addInts := func(a, b int32) int32 { return a + b }
	concatStrings := func(s1, s2 string) string { return s1 + s2 }
	bytesLen := func(b []byte) int64 { return int64(len(b)) }
	timestampYear := func(ts time.Time) int32 { return int32(ts.Year()) }
	isTrue := func(b bool) bool { return b }
	addFloats := func(a, b float64) float64 { return a + b }
	divZeroPanic := func(a int32) int32 {
		zero := int32(0)
		return a / zero
	}
	customPanicFunc := func() int32 { panic("custom panic from UDF") }

	// UDFs for struct and map testing
	getStructName := func(s MyStruct) string { return s.Name }
	updateStructVal := func(s MyStruct, increment float64) MyStruct {
		s.Val += increment
		return s
	}
	getNestedStructTag := func(ns MyNestedStruct) string { return ns.Tag }
	getMapValue := func(m map[string]int32, key string) int32 {
		val, ok := m[key]
		if !ok {
			return -1
		} // Indicate key not found
		return val
	}
	addtoMap := func(m map[string]string, k, v string) map[string]string {
		if m == nil {
			m = make(map[string]string)
		}
		m[k] = v
		return m
	}
	// Modified structToMap to return a JSON string
	structToJsonString := func(s MyStruct) string {
		Map := map[string]any{"ID": s.ID, "Name": s.Name, "Val": s.Val}
		jsonBytes, err := json.Marshal(Map)
		if err != nil {
			panic(fmt.Sprintf("json.Marshal failed in structToJsonString UDF: %v", err))
		}
		return string(jsonBytes)
	}

	// New Variadic UDFs
	sumIntsVariadic := func(label string, numbers ...int32) string {
		var sum int32 = 0
		for _, num := range numbers {
			sum += num
		}
		return fmt.Sprintf("%s: %d", label, sum)
	}
	concatStringsVariadic := func(separator string, parts ...string) string {
		return strings.Join(parts, separator)
	}
	countVariadicArgs := func(fixed bool, variadicArgs ...float64) int64 {
		return int64(len(variadicArgs))
	}
	sumAllDoublesVariadic := func(values ...float64) float64 {
		var sum float64 = 0
		for _, v := range values {
			sum += v
		}
		return sum
	}

	// UDF for testing special null handling
	handleNilString := func(s *string) string {
		if s == nil {
			return "SQL NULL was received as nil"
		}
		return *s
	}

	tests := []struct {
		name          string
		udfName       string
		goFunc        any
		options       []func(*udfOption)
		query         string
		prepareParams func(t *testing.T) []interface{}
		expectedValue interface{}
		expectError   bool
		errorContains string
	}{
		// Basic types and existing error tests (queryParams defined directly)
		{
			name: "add ints", udfName: "add_ints_udf", goFunc: addInts,
			options:       nil,
			query:         "SELECT add_ints_udf(CAST(? AS INTEGER), CAST(? AS INTEGER))",
			prepareParams: func(t *testing.T) []interface{} { return []interface{}{5, 10} },
			expectedValue: int64(15),
		},
		{
			name: "add ints volatile", udfName: "add_ints_volatile_udf", goFunc: addInts,
			options:       []func(*udfOption){WithVolatile(true)},
			query:         "SELECT add_ints_volatile_udf(CAST(? AS INTEGER), CAST(? AS INTEGER))",
			prepareParams: func(t *testing.T) []interface{} { return []interface{}{5, 10} },
			expectedValue: int64(15),
		},
		{
			name: "concat strings", udfName: "concat_str_udf", goFunc: concatStrings,
			options:       nil,
			query:         "SELECT concat_str_udf(?, ?)",
			prepareParams: func(t *testing.T) []interface{} { return []interface{}{"hello ", "world"} },
			expectedValue: "hello world",
		},
		{
			name: "bytes length", udfName: "bytes_len_udf", goFunc: bytesLen,
			options:       nil,
			query:         "SELECT bytes_len_udf(?)",
			prepareParams: func(t *testing.T) []interface{} { return []interface{}{[]byte{0x01, 0x02, 0x03}} },
			expectedValue: int64(3),
		},
		{
			name: "timestamp year", udfName: "ts_year_udf", goFunc: timestampYear,
			options:       nil,
			query:         "SELECT ts_year_udf(?)",
			prepareParams: func(t *testing.T) []interface{} { return []interface{}{time.Date(2024, 7, 15, 0, 0, 0, 0, time.UTC)} },
			expectedValue: int64(2024),
		},
		{
			name: "boolean passthrough", udfName: "is_true_udf", goFunc: isTrue,
			options:       nil,
			query:         "SELECT is_true_udf(?)",
			prepareParams: func(t *testing.T) []interface{} { return []interface{}{true} },
			expectedValue: true,
		},
		{
			name: "add floats", udfName: "add_floats_udf", goFunc: addFloats,
			options:       nil,
			query:         "SELECT add_floats_udf(?, ?)",
			prepareParams: func(t *testing.T) []interface{} { return []interface{}{1.2, 3.4} },
			expectedValue: float64(4.6),
		},
		{
			name: "udf division by zero", udfName: "div_zero_udf", goFunc: divZeroPanic,
			options:       nil,
			query:         "SELECT div_zero_udf(CAST(? AS INTEGER))",
			prepareParams: func(t *testing.T) []interface{} { return []interface{}{10} },
			expectError:   true, errorContains: "panic in UDF (func type func(int32) int32): runtime error: integer divide by zero",
		},
		{
			name: "udf custom panic", udfName: "custom_panic_udf", goFunc: customPanicFunc,
			options:       nil,
			query:         "SELECT custom_panic_udf()",
			prepareParams: func(t *testing.T) []interface{} { return []interface{}{} },
			expectError:   true, errorContains: "panic in UDF (func type func() int32): custom panic from UDF",
		},

		// Struct UDF Tests - using JSON string cast to STRUCT for parameters
		{
			name: "get struct name", udfName: "get_s_name_udf", goFunc: getStructName,
			options: nil,
			query:   "SELECT get_s_name_udf(CAST(? AS STRUCT(ID INTEGER, Name VARCHAR, Val DOUBLE)))",
			prepareParams: func(t *testing.T) []interface{} {
				jsonData, err := json.Marshal(map[string]interface{}{"ID": int32(1), "Name": "StructA", "Val": 1.23})
				if err != nil {
					t.Fatalf("json.Marshal failed: %v", err)
				}
				return []interface{}{string(jsonData)}
			},
			expectedValue: "StructA",
		},
		{
			name: "update struct val and return struct", udfName: "upd_s_val_udf", goFunc: updateStructVal,
			options: nil,
			query:   "SELECT upd_s_val_udf(CAST(? AS STRUCT(ID INTEGER, Name VARCHAR, Val DOUBLE)), ?)",
			prepareParams: func(t *testing.T) []interface{} {
				jsonData, err := json.Marshal(map[string]interface{}{"ID": int32(2), "Name": "StructB", "Val": 2.5})
				if err != nil {
					t.Fatalf("json.Marshal failed: %v", err)
				}
				return []interface{}{string(jsonData), 1.5}
			},
			expectedValue: map[string]interface{}{"ID": int32(2), "Name": "StructB", "Val": float64(4.0)},
		},
		{
			name: "get nested struct tag", udfName: "get_ns_tag_udf", goFunc: getNestedStructTag,
			options: nil,
			query:   "SELECT get_ns_tag_udf(CAST(? AS STRUCT(Tag VARCHAR, Nested STRUCT(ID INTEGER, Name VARCHAR, Val DOUBLE))))",
			prepareParams: func(t *testing.T) []interface{} {
				jsonData, err := json.Marshal(map[string]interface{}{
					"Tag":    "OuterTag",
					"Nested": map[string]interface{}{"ID": int32(3), "Name": "InnerStruct", "Val": 3.14},
				})
				if err != nil {
					t.Fatalf("json.Marshal failed: %v", err)
				}
				return []interface{}{string(jsonData)}
			},
			expectedValue: "OuterTag",
		},

		// Map UDF Tests (Map parameters are generally well-supported by DuckDB with map literals)
		{
			name: "get map value", udfName: "get_map_val_udf", goFunc: getMapValue,
			options:       nil,
			query:         "SELECT get_map_val_udf(map {'a': 10, 'b': 20}, ?)",
			prepareParams: func(t *testing.T) []interface{} { return []interface{}{"b"} },
			expectedValue: int64(20),
		},
		{
			name: "get map value (key not found)", udfName: "get_map_val_nf_udf", goFunc: getMapValue,
			options:       nil,
			query:         "SELECT get_map_val_nf_udf(map {'x': 1, 'y': 2}, ?)",
			prepareParams: func(t *testing.T) []interface{} { return []interface{}{"z"} },
			expectedValue: int64(-1),
		},
		{
			name: "add to map and return map", udfName: "add_to_map_udf", goFunc: addtoMap,
			options:       nil,
			query:         "SELECT add_to_map_udf(map {'first': 'apple'}, ?, ?)",
			prepareParams: func(t *testing.T) []interface{} { return []interface{}{"second", "banana"} },
			expectedValue: duckdb.Map(map[any]any{"first": "apple", "second": "banana"}),
		},
		{
			name: "struct to JSON string UDF", udfName: "s_to_json_udf", goFunc: structToJsonString,
			options: nil,
			query:   "SELECT s_to_json_udf(CAST(? AS STRUCT(ID INTEGER, Name VARCHAR, Val DOUBLE)))",
			prepareParams: func(t *testing.T) []interface{} {
				jsonData, err := json.Marshal(map[string]interface{}{"ID": int32(7), "Name": "TomTom", "Val": 7.7})
				if err != nil {
					t.Fatalf("json.Marshal failed: %v", err)
				}
				return []interface{}{string(jsonData)}
			},
			expectedValue: "{\"ID\":7,\"Name\":\"TomTom\",\"Val\":7.7}",
		},

		// Variadic UDF Tests
		{
			name: "sum ints variadic (multiple)", udfName: "sum_ints_var_udf", goFunc: sumIntsVariadic,
			options:       nil,
			query:         "SELECT sum_ints_var_udf(?, CAST(? AS INTEGER), CAST(? AS INTEGER), CAST(? AS INTEGER))",
			prepareParams: func(t *testing.T) []interface{} { return []interface{}{"SumX", 10, 20, 30} },
			expectedValue: "SumX: 60",
		},
		{
			name: "sum ints variadic (zero variadic)", udfName: "sum_ints_var_zero_udf", goFunc: sumIntsVariadic,
			options:       nil,
			query:         "SELECT sum_ints_var_zero_udf(?)",
			prepareParams: func(t *testing.T) []interface{} { return []interface{}{"SumY"} },
			expectedValue: "SumY: 0",
		},
		{
			name: "concat strings variadic (multiple)", udfName: "cat_strs_var_udf", goFunc: concatStringsVariadic,
			options:       nil,
			query:         "SELECT cat_strs_var_udf(?, ?, ?, ?)",
			prepareParams: func(t *testing.T) []interface{} { return []interface{}{"-", "a", "b", "c"} },
			expectedValue: "a-b-c",
		},
		{
			name: "concat strings variadic (zero variadic)", udfName: "cat_strs_var_zero_udf", goFunc: concatStringsVariadic,
			options:       nil,
			query:         "SELECT cat_strs_var_zero_udf(?)",
			prepareParams: func(t *testing.T) []interface{} { return []interface{}{","} },
			expectedValue: "", // strings.Join with empty slice results in empty string
		},
		{
			name: "count variadic args (three)", udfName: "count_var_udf", goFunc: countVariadicArgs,
			options:       nil,
			query:         "SELECT count_var_udf(?, ?, ?, ?)",
			prepareParams: func(t *testing.T) []interface{} { return []interface{}{true, 1.1, 2.2, 3.3} },
			expectedValue: int64(3),
		},
		{
			name: "count variadic args (zero)", udfName: "count_var_zero_udf", goFunc: countVariadicArgs,
			options:       nil,
			query:         "SELECT count_var_zero_udf(?)",
			prepareParams: func(t *testing.T) []interface{} { return []interface{}{false} },
			expectedValue: int64(0),
		},
		{
			name: "sum all doubles variadic (multiple)", udfName: "sum_all_f64_udf", goFunc: sumAllDoublesVariadic,
			options:       nil,
			query:         "SELECT sum_all_f64_udf(?, ?, ?)",
			prepareParams: func(t *testing.T) []interface{} { return []interface{}{1.5, 2.5, 3.0} },
			expectedValue: float64(7.0),
		},
		{
			name: "sum all doubles variadic (zero)", udfName: "sum_all_f64_zero_udf", goFunc: sumAllDoublesVariadic,
			options:       nil,
			query:         "SELECT sum_all_f64_zero_udf()",
			prepareParams: func(t *testing.T) []interface{} { return []interface{}{} },
			expectedValue: float64(0),
		},
		// Special Null Handling Tests
		{
			name: "special null handling (input NULL)", udfName: "handle_nil_str_udf", goFunc: handleNilString,
			options:       []func(*udfOption){WithSpecialNullHandling(true)},
			query:         "SELECT handle_nil_str_udf(NULL)",
			prepareParams: nil,
			expectedValue: "SQL NULL was received as nil",
		},
		{
			name: "special null handling (input non-NULL)", udfName: "handle_nil_str_non_null_udf", goFunc: handleNilString,
			options:       []func(*udfOption){WithSpecialNullHandling(true)},
			query:         "SELECT handle_nil_str_non_null_udf(?)",
			prepareParams: func(t *testing.T) []interface{} { return []interface{}{"test string"} },
			expectedValue: "test string",
		},
		{
			name: "default null handling (input NULL, UDF not called)", udfName: "handle_nil_str_default_udf", goFunc: handleNilString,
			options:       nil,
			query:         "SELECT handle_nil_str_default_udf(NULL)",
			prepareParams: nil,
			expectedValue: nil, // Expect SQL NULL to propagate directly
		},
	}

	// Main DB connection for the test suite, from which specific connections are derived
	db, err := sql.Open("duckdb", "")
	if err != nil {
		t.Fatalf("Failed to open DuckDB for test suite: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping DuckDB for test suite: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn, err := db.Conn(context.Background())
			if err != nil {
				t.Fatalf("Failed to get dedicated DB connection for UDF test '%s': %v", tt.name, err)
			}
			defer conn.Close()

			udfImpl, err := BuildScalarUDF(tt.goFunc, tt.options...)
			if err != nil {
				t.Fatalf("Failed to build UDF '%s': %v", tt.udfName, err)
			}

			err = duckdb.RegisterScalarUDF(conn, tt.udfName, udfImpl)
			if err != nil {
				t.Fatalf("Failed to register UDF '%s' on dedicated connection: %v", tt.udfName, err)
			}

			var queryParams []interface{}
			if tt.prepareParams != nil {
				queryParams = tt.prepareParams(t)
			}

			if tt.expectError {
				expectQueryErrorOnConn(t, conn, tt.errorContains, tt.query, queryParams...)
			} else {
				result := querySingleValueOnConn(t, conn, tt.query, queryParams...)
				// For nil expected values, check if result is nil
				if tt.expectedValue == nil {
					assertNil(t, result, "Expected nil but got %v", result)
				} else {
					// For []byte, use special comparison
					if expBytes, ok := tt.expectedValue.([]byte); ok {
						actBytes, ok := result.([]byte)
						assertTrue(t, ok, "Expected []byte but got %T", result)
						assertEqual(t, expBytes, actBytes, "Byte arrays don't match")
					} else if expTime, ok := tt.expectedValue.(time.Time); ok {
						// For time.Time, compare with microsecond precision
						actTime, ok := result.(time.Time)
						assertTrue(t, ok, "Expected time.Time but got %T", result)
						assertTrue(t, expTime.In(time.UTC).Truncate(time.Microsecond).Equal(actTime.In(time.UTC).Truncate(time.Microsecond)),
							"Times don't match: expected %v, got %v", expTime, actTime)
					} else if expInt64, ok := tt.expectedValue.(int64); ok {
						// Special handling for int64 vs int32 comparison
						switch v := result.(type) {
						case int32:
							assertEqual(t, expInt64, int64(v), "Integer values don't match")
						case int64:
							assertEqual(t, expInt64, v, "Integer values don't match")
						default:
							fail(t, "Expected int32 or int64 but got %T", result)
						}
					} else {
						// For other types, use deep equal
						assertEqual(t, tt.expectedValue, result, "Result doesn't match expected value")
					}
				}
			}
		})
	}
}
