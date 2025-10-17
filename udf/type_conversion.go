package udf

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"time"

	"github.com/duckdb/duckdb-go/v2"
)

// goTypeToDuckDBTypeInfo converts a Go reflect.Type to a DuckDB TypeInfo.
// This function returns duckdb.TypeInfo, which encapsulates the logical description of the type.
//
// Supported Go types include:
// - Basic types: int/int8/int16/int32 -> INTEGER, int64 -> BIGINT
// - Unsigned integers: uint/uint8/uint16/uint32 -> UINTEGER, uint64 -> UBIGINT
// - Floating-point: float32 -> FLOAT, float64 -> DOUBLE
// - String: string -> VARCHAR
// - Boolean: bool -> BOOLEAN
// - Binary data: []byte -> BLOB
// - Time: time.Time -> TIMESTAMP
// - Struct: struct -> STRUCT (only exported fields are considered)
// - Map: map[K]V -> MAP (K and V must be supported types)
//
// Pointer types (like *struct, *map, *time.Time) will be automatically dereferenced once.
// Unsupported types include channels, functions, interfaces, and slices (except []byte).
// Empty structs (with no exported fields) are also not supported.
func goTypeToDuckDBTypeInfo(rt reflect.Type) (duckdb.TypeInfo, error) {
	// Handle pointer dereferencing.
	if rt.Kind() == reflect.Pointer {
		elemType := rt.Elem()
		// Dereference if the element is a struct, map, time.Time, or a supported basic type.
		switch elemType.Kind() {
		case reflect.Struct, reflect.Map:
			rt = elemType // Dereference for struct and map
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64,
			reflect.String,
			reflect.Bool:
			rt = elemType // Dereference for supported basic types as well
		default:
			// For other pointer types (e.g., pointer to slice, pointer to unsupported type),
			// check if it's *time.Time specifically, otherwise it might be unsupported.
			if elemType.PkgPath() != "time" || elemType.Name() != "Time" {
				// Fall through to the main switch for default error handling if not *time.Time
				// or if it's a pointer to something we don't want to auto-dereference here.
			} else {
				rt = elemType // Dereference *time.Time
			}
		}
	}

	// Specific check for time.Time MUST come BEFORE the general reflect.Struct case
	if rt.PkgPath() == "time" && rt.Name() == "Time" {
		duckDBAPIType := duckdb.TYPE_TIMESTAMP
		typeInfo, err := duckdb.NewTypeInfo(duckDBAPIType)
		if err != nil {
			return nil, fmt.Errorf("error creating TypeInfo for DuckDB TIMESTAMP (from Go type %s): %w", rt.String(), err)
		}
		return typeInfo, nil
	}

	var duckDBAPIType duckdb.Type

	switch rt.Kind() {
	case reflect.Int, reflect.Int16, reflect.Int8, reflect.Int32:
		duckDBAPIType = duckdb.TYPE_INTEGER
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint8:
		// Use DuckDB's unsigned types instead of signed types to avoid overflow issues
		duckDBAPIType = duckdb.TYPE_UINTEGER
	case reflect.Int64:
		duckDBAPIType = duckdb.TYPE_BIGINT
	case reflect.Uint64:
		// Use DuckDB's unsigned BIGINT type
		duckDBAPIType = duckdb.TYPE_UBIGINT
	case reflect.Float32:
		duckDBAPIType = duckdb.TYPE_FLOAT
	case reflect.Float64:
		duckDBAPIType = duckdb.TYPE_DOUBLE
	case reflect.String:
		duckDBAPIType = duckdb.TYPE_VARCHAR
	case reflect.Bool:
		duckDBAPIType = duckdb.TYPE_BOOLEAN
	case reflect.Slice:
		if rt.Elem().Kind() == reflect.Uint8 { // []byte
			duckDBAPIType = duckdb.TYPE_BLOB
		} else {
			return nil, fmt.Errorf("unsupported Go slice element type for UDF: %s (only []byte for BLOB)", rt.Elem().String())
		}
	case reflect.Struct:
		var structEntries []duckdb.StructEntry
		for i := 0; i < rt.NumField(); i++ {
			field := rt.Field(i)
			if !field.IsExported() { // PkgPath is empty for exported fields
				continue
			}
			fieldName := field.Name // Use Go field name directly for DuckDB struct field name
			fieldTypeInfo, err := goTypeToDuckDBTypeInfo(field.Type)
			if err != nil {
				return nil, fmt.Errorf("error converting field '%s' of struct %s: %w", fieldName, rt.Name(), err)
			}
			entry, err := duckdb.NewStructEntry(fieldTypeInfo, fieldName)
			if err != nil {
				return nil, fmt.Errorf("error creating struct entry for field '%s' of struct %s: %w", fieldName, rt.Name(), err)
			}
			structEntries = append(structEntries, entry)
		}
		if len(structEntries) == 0 {
			return nil, fmt.Errorf("cannot create DuckDB STRUCT from Go struct %s with no exported fields", rt.Name())
		}
		// duckdb.NewStructInfo takes (first StructEntry, others ...StructEntry)
		var finalStructInfo duckdb.TypeInfo
		var err error
		if len(structEntries) == 1 {
			finalStructInfo, err = duckdb.NewStructInfo(structEntries[0])
		} else {
			finalStructInfo, err = duckdb.NewStructInfo(structEntries[0], structEntries[1:]...)
		}
		if err != nil {
			return nil, fmt.Errorf("error creating StructInfo for %s: %w", rt.Name(), err)
		}
		return finalStructInfo, nil
	case reflect.Map:
		keyType := rt.Key()
		valType := rt.Elem()

		keyTypeInfo, err := goTypeToDuckDBTypeInfo(keyType)
		if err != nil {
			return nil, fmt.Errorf("error converting map key type %s for UDF: %w", keyType.String(), err)
		}
		// TODO: Add validation for DuckDB map key types (e.g., not complex types themselves)
		// For now, assume goTypeToDuckDBTypeInfo handles basic unsupported types.

		valTypeInfo, err := goTypeToDuckDBTypeInfo(valType)
		if err != nil {
			return nil, fmt.Errorf("error converting map value type %s for UDF: %w", valType.String(), err)
		}

		mapInfo, err := duckdb.NewMapInfo(keyTypeInfo, valTypeInfo)
		if err != nil {
			return nil, fmt.Errorf("error creating MapInfo for map[%s]%s: %w", keyType.String(), valType.String(), err)
		}
		return mapInfo, nil
	case reflect.Interface:
		// interface{} or any is too generic to map to a specific DuckDB type for TypeInfo.
		// DuckDB needs concrete types for schema and UDF signatures.
		return nil, fmt.Errorf("unsupported Go type: interface {} (any) is not directly mappable to a concrete DuckDB type for UDF signature. Consider using specific types or structs/maps with specific field/value types")
	default:
		// All other kinds (Chan, Func, Ptr to unhandled type, etc.) are unsupported here.
		// Note: Ptr was handled at the beginning of the function for specific cases (struct, map, time.Time).
		// If a Ptr to a primitive or other unsupported type reaches here, it means it wasn't dereferenced or handled.
		return nil, fmt.Errorf("unsupported Go type kind for UDF: %s (specific type: %s)", rt.Kind().String(), rt.String())
	}

	if duckDBAPIType == 0 { // duckDBAPIType not set validly
		return nil, fmt.Errorf("failed to map Go type %s to a DuckDB API type (duckDBAPIType is zero)", rt.String())
	}

	// Use duckdb.NewTypeInfo to create a TypeInfo instance
	// NewTypeInfo takes a duckdb.Type (which is an alias for mapping.Type)
	typeInfo, err := duckdb.NewTypeInfo(duckDBAPIType)
	if err != nil {
		return nil, fmt.Errorf("error creating TypeInfo for DuckDB type %v (from Go type %s): %w", duckDBAPIType, rt.String(), err)
	}
	return typeInfo, nil
}

// convertToReflectValue converts a value from DuckDB (via driver.Value) to a reflect.Value expected by the user function.
//
// This function supports the following type conversions:
// - SQL NULL -> Go nil (for pointer, slice, map, channel, function, and interface types)
// - SQL NULL -> Go zero value struct (when target is a struct)
// - SQL integer types -> Go integer types (int, int8-64, uint, uint8-64)
// - SQL floating-point types -> Go floating-point types (float32, float64)
// - SQL string -> Go string
// - SQL boolean -> Go boolean
// - SQL BLOB -> Go []byte
// - SQL TIMESTAMP (microsecond value) -> Go time.Time
// - SQL STRUCT -> Go struct (field names must match)
// - SQL MAP -> Go map (key and value types must match)
//
// Special restrictions:
// - No implicit numeric to string conversion allowed (prevents unexpected data loss)
// - No implicit numeric to boolean conversion allowed
// - Struct conversion requires all exported fields of the target struct to have corresponding values in the source map
// - Map conversion requires key and value types that can be converted to the target map's key and value types
func convertToReflectValue(sourceVal driver.Value, targetType reflect.Type) (reflect.Value, error) {
	if sourceVal == nil {
		// If the target is a pointer, slice, map, chan, func, or interface, a nil sourceVal maps to a nil reflect.Value of that type.
		// For structs, it maps to a zero struct.
		switch targetType.Kind() {
		case reflect.Pointer, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func, reflect.Interface:
			return reflect.Zero(targetType), nil
		case reflect.Struct: // Allow nil to be converted to a zero struct if target is a struct
			return reflect.Zero(targetType), nil
		default:
			return reflect.Value{}, fmt.Errorf("cannot convert nil (driver.Value) to Go type %s for UDF parameter", targetType.String())
		}
	}

	// Handle time.Time specifically and early, as its Kind is Struct but needs special conversion from int64 (micros) or other specific types.
	if targetType.PkgPath() == "time" && targetType.Name() == "Time" {
		switch sv := sourceVal.(type) {
		case int64: // Unix microseconds are common for DuckDB timestamps
			return reflect.ValueOf(time.UnixMicro(sv).In(time.UTC)), nil
		case string:
			// Add support for string to time.Time conversion
			// Try multiple common time formats
			formats := []string{
				time.RFC3339,
				"2006-01-02 15:04:05",
				"2006-01-02T15:04:05",
				"2006-01-02",
				time.RFC3339Nano,
			}

			for _, format := range formats {
				if t, err := time.Parse(format, sv); err == nil {
					return reflect.ValueOf(t), nil
				}
			}
			return reflect.Value{}, fmt.Errorf("unable to parse string '%s' as time.Time, tried multiple common formats", sv)
		case time.Time: // If source is already time.Time (e.g. from another Go UDF or driver optimization)
			return reflect.ValueOf(sv), nil
		default:
			return reflect.Value{}, fmt.Errorf("cannot convert source type %T to Go type time.Time for UDF parameter", sourceVal)
		}
	}

	sourceReflectVal := reflect.ValueOf(sourceVal)

	if sourceReflectVal.Type().AssignableTo(targetType) {
		return sourceReflectVal, nil
	}

	// Handle specific target kinds before general convertibility for more control
	switch targetType.Kind() {
	case reflect.Struct: // time.Time is already handled
		srcMap, ok := sourceVal.(map[string]any)
		if !ok {
			return reflect.Value{}, fmt.Errorf("expected map[string]interface{} from driver for DuckDB STRUCT, but got %T for target Go struct %s", sourceVal, targetType.Name())
		}
		newStruct := reflect.New(targetType).Elem()
		for i := 0; i < targetType.NumField(); i++ {
			fieldSpec := targetType.Field(i)
			if !fieldSpec.IsExported() {
				continue
			}
			valFromMap, exists := srcMap[fieldSpec.Name]
			if !exists {
				return reflect.Value{}, fmt.Errorf("field '%s' missing in source map from DuckDB STRUCT for target Go struct %s", fieldSpec.Name, targetType.Name())
			}
			convertedFieldVal, err := convertToReflectValue(valFromMap, fieldSpec.Type)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("error converting field '%s' for Go struct %s: %w", fieldSpec.Name, targetType.Name(), err)
			}
			newStruct.Field(i).Set(convertedFieldVal)
		}
		return newStruct, nil

	case reflect.Map:
		srcDuckDBMap, ok := sourceVal.(duckdb.Map)
		if !ok {
			return reflect.Value{}, fmt.Errorf("expected duckdb.Map from driver for DuckDB MAP, but got %T for target Go map %s", sourceVal, targetType.String())
		}
		goMapType := targetType
		newGoMap := reflect.MakeMapWithSize(goMapType, len(srcDuckDBMap))
		goMapKeyType := goMapType.Key()
		goMapElemType := goMapType.Elem()
		for k, v := range srcDuckDBMap {
			convertedKey, err := convertToReflectValue(k, goMapKeyType)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("error converting map key for target Go map %s: %w", goMapType.String(), err)
			}

			// Check if key type matches before processing map values
			// This ensures key type mismatch errors take precedence over value type mismatch errors
			if !convertedKey.Type().AssignableTo(goMapKeyType) {
				return reflect.Value{}, fmt.Errorf("error converting map key: %v (type %T) is not assignable to map key type %s",
					k, k, goMapKeyType.String())
			}

			convertedValue, err := convertToReflectValue(v, goMapElemType)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("error converting map value for key '%v' for target Go map %s: %w",
					convertedKey.Interface(), goMapType.String(), err)
			}

			newGoMap.SetMapIndex(convertedKey, convertedValue)
		}
		return newGoMap, nil
	}

	// Handle cases where targetType is a pointer to a basic type (e.g. *string, *int)
	if targetType.Kind() == reflect.Pointer {
		elemType := targetType.Elem()
		// Attempt to convert sourceVal to the element type first
		convertedElemVal, err := convertToReflectValue(sourceVal, elemType)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("error converting to element type %s for pointer target %s: %w", elemType.String(), targetType.String(), err)
		}
		// Create a new pointer to the converted element value
		newPtrVal := reflect.New(elemType)
		newPtrVal.Elem().Set(convertedElemVal)
		return newPtrVal, nil
	}

	// General convertibility (after specific struct/map and time.Time handling)
	if sourceReflectVal.Type().ConvertibleTo(targetType) {
		// Disallow certain implicit type conversions that might lead to data loss or misinterpretation
		if targetType.Kind() == reflect.String && sourceReflectVal.Kind() != reflect.String {
			// Disallow implicit number to string conversion as this is typically unintentional and should use explicit formatting
			return reflect.Value{}, fmt.Errorf("cannot convert DuckDB driver.Value of type %s (Go type %s, value: %v) to Go function parameter type string",
				reflect.TypeOf(sourceVal).String(), sourceReflectVal.Type().String(), sourceVal)
		} else if targetType.Kind() == reflect.Bool && sourceReflectVal.Kind() != reflect.Bool {
			// Disallow implicit number to boolean conversion
			return reflect.Value{}, fmt.Errorf("cannot convert DuckDB driver.Value of type %s (Go type %s, value: %v) to Go function parameter type bool",
				reflect.TypeOf(sourceVal).String(), sourceReflectVal.Type().String(), sourceVal)
		} else {
			return sourceReflectVal.Convert(targetType), nil
		}
	}

	// Fallback error: if time.Time was the target but sourceVal wasn't int64/string/time.Time, it would have errored in the time.Time block.
	// This error is for all other unhandled conversions.
	return reflect.Value{}, fmt.Errorf("cannot convert DuckDB driver.Value of type %s (Go type %s, value: %v) to Go function parameter type %s",
		reflect.TypeOf(sourceVal).String(), sourceReflectVal.Type().String(), sourceVal, targetType.String())
}
