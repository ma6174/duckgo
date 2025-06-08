// Package udf provides Go implementation support for DuckDB User Defined Functions (UDF).
//
// This package allows you to create DuckDB scalar UDFs using Go functions, supporting various types including basic types
// (integers, floats, strings, booleans), complex types (structs, maps), and special types (time, binary data).
// It also supports variadic functions and special NULL handling.
//
// # Basic Usage
//
// Use BuildScalarUDF to create a UDF, then register it to a DuckDB connection using duckdb.RegisterScalarUDF:
//
//	// Define a simple Go function
//	add := func(a, b int) int { return a + b }
//
//	// Build the UDF
//	udfImpl, err := udf.BuildScalarUDF(add)
//	if err != nil {
//		// Handle error
//	}
//
//	// Register the UDF to a DuckDB connection
//	err = duckdb.RegisterScalarUDF(conn, "add_func", udfImpl)
//	if err != nil {
//		// Handle error
//	}
//
//	// Now you can use this function in SQL
//	// SELECT add_func(5, 10) -- returns 15
//
// # Supported Types
//
// UDFs support the following Go types as parameters and return values:
//
// - Integer types: int/int8/int16/int32/int64/uint/uint8/uint16/uint32/uint64
// - Floating-point types: float32/float64
// - String: string
// - Boolean: bool
// - Binary data: []byte
// - Time: time.Time
// - Struct: struct (must have exported fields)
// - Map: map[K]V (K and V must be supported types)
// - Pointers: *struct, *map[K]V, *time.Time (will be automatically dereferenced)
//
// # Special Features
//
// This package supports several special features:
//
// 1. Variadic functions:
//
//	// UDF with variadic parameters
//	sum := func(numbers ...int) int {
//		total := 0
//		for _, n := range numbers {
//			total += n
//		}
//		return total
//	}
//
// 2. Special NULL handling (using pointer parameters):
//
//	// Use WithSpecialNullHandling option to enable special NULL handling
//	handleNull := func(s *string) string {
//		if s == nil { // SQL NULL values are converted to Go nil
//			return "NULL"
//		}
//		return *s
//	}
//	udfImpl, _ := udf.BuildScalarUDF(handleNull, udf.WithSpecialNullHandling(true))
//
// 3. Volatile functions (must recalculate each call, results cannot be cached):
//
//	// Use WithVolatile option to mark a function as volatile
//	getTime := func() time.Time { return time.Now() }
//	udfImpl, _ := udf.BuildScalarUDF(getTime, udf.WithVolatile(true))
//
// # Error Handling
//
// Panics during UDF execution are caught and converted to SQL errors with detailed context information.
// Errors during UDF building and registration also return detailed error messages.
package udf
