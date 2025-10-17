package udf

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"runtime"

	"github.com/duckdb/duckdb-go/v2"
)

// autoScalarFunc is an internal struct that implements the duckdb.ScalarFunc interface.
// It wraps a user-provided Go function and handles type conversion and configuration required for DuckDB UDF registration.
type autoScalarFunc struct {
	userFunc             reflect.Value
	goArgTypes           []reflect.Type // If variadic function, the last type is the slice type e.g. []int for (...int)
	goReturnType         reflect.Type
	duckDBInputTypeInfos []duckdb.TypeInfo // Only contains TypeInfo for fixed parameters
	duckDBResultTypeInfo duckdb.TypeInfo
	specialNullHandling  bool
	volatile             bool

	isVariadic             bool            // Flag indicating if this is a variadic UDF
	duckDBVariadicTypeInfo duckdb.TypeInfo // TypeInfo for the variadic part (based on element type)
}

// Config method adjusted to set VariadicTypeInfo
func (asf *autoScalarFunc) Config() duckdb.ScalarFuncConfig {
	conf := duckdb.ScalarFuncConfig{
		InputTypeInfos:      asf.duckDBInputTypeInfos, // These are only for fixed parameters
		ResultTypeInfo:      asf.duckDBResultTypeInfo,
		SpecialNullHandling: asf.specialNullHandling,
		Volatile:            asf.volatile,
	}
	if asf.isVariadic {
		conf.VariadicTypeInfo = asf.duckDBVariadicTypeInfo
	}
	return conf
}

// Helper function to process variadic arguments
func (asf *autoScalarFunc) processVariadicArgs(inputArgs []driver.Value, numFixedGoParams int) ([]reflect.Value, error) {
	// Total number of formal parameters in the function signature
	numFormalGoParams := len(asf.goArgTypes)

	// Validate argument count
	if len(inputArgs) < numFixedGoParams {
		return nil, fmt.Errorf("UDF (variadic, func type %s) requires at least %d fixed parameters, but only %d were provided",
			asf.userFunc.Type().String(), numFixedGoParams, len(inputArgs))
	}

	// Calculate the total number of arguments needed for reflect.Call()
	numVariadicInputsProvided := len(inputArgs) - numFixedGoParams
	numCallArgs := numFixedGoParams + numVariadicInputsProvided
	callArgs := make([]reflect.Value, numCallArgs)

	// Convert fixed parameters
	for i := range numFixedGoParams {
		goArgType := asf.goArgTypes[i] // Type of the i-th fixed Go parameter
		duckDBVal := inputArgs[i]
		convertedVal, conversionErr := convertToReflectValue(duckDBVal, goArgType)
		if conversionErr != nil {
			return nil, fmt.Errorf("error converting fixed parameter %d (Go type %s, func type %s): %w",
				i, goArgType.String(), asf.userFunc.Type().String(), conversionErr)
		}
		callArgs[i] = convertedVal
	}

	// Convert and append variadic parameters
	// Get the variadic slice type from the function signature (e.g., reflect.TypeOf([]int))
	variadicGoSliceType := asf.goArgTypes[numFormalGoParams-1] // Type of the last formal parameter
	variadicGoElemType := variadicGoSliceType.Elem()           // Element type (e.g., reflect.TypeOf(int))

	for i := range numVariadicInputsProvided {
		duckDBVal := inputArgs[numFixedGoParams+i]
		convertedVal, conversionErr := convertToReflectValue(duckDBVal, variadicGoElemType)
		if conversionErr != nil {
			return nil, fmt.Errorf("error converting variadic parameter %d (overall input arg %d, Go element type %s, func type %s): %w",
				i, numFixedGoParams+i, variadicGoElemType.String(), asf.userFunc.Type().String(), conversionErr)
		}
		callArgs[numFixedGoParams+i] = convertedVal
	}

	return callArgs, nil
}

// Helper function to process non-variadic arguments
func (asf *autoScalarFunc) processNonVariadicArgs(inputArgs []driver.Value) ([]reflect.Value, error) {
	numFormalGoParams := len(asf.goArgTypes)

	// Validate argument count
	if len(inputArgs) != numFormalGoParams {
		return nil, fmt.Errorf("UDF (non-variadic, func type %s) requires %d parameters, but %d were provided",
			asf.userFunc.Type().String(), numFormalGoParams, len(inputArgs))
	}

	// Convert all parameters
	callArgs := make([]reflect.Value, numFormalGoParams)
	for i := range numFormalGoParams {
		goArgType := asf.goArgTypes[i]
		duckDBVal := inputArgs[i]
		convertedVal, conversionErr := convertToReflectValue(duckDBVal, goArgType)
		if conversionErr != nil {
			return nil, fmt.Errorf("error converting parameter %d (Go type %s, func type %s): %w",
				i, goArgType.String(), asf.userFunc.Type().String(), conversionErr)
		}
		callArgs[i] = convertedVal
	}

	return callArgs, nil
}

// Executor().RowExecutor method simplified to use the new helper functions
func (asf *autoScalarFunc) Executor() duckdb.ScalarFuncExecutor {
	return duckdb.ScalarFuncExecutor{
		RowExecutor: func(inputArgs []driver.Value) (result any, err error) {
			defer func() {
				if r := recover(); r != nil {
					// Provide richer error context, including function type, parameter info, and stack trace
					argValues := make([]string, len(inputArgs))
					for i, arg := range inputArgs {
						if arg == nil {
							argValues[i] = "NULL"
						} else {
							argValues[i] = fmt.Sprintf("%v (type %T)", arg, arg)
						}
					}

					// Get stack trace
					buf := make([]byte, 4096)
					n := runtime.Stack(buf, false)
					stackTrace := string(buf[:n])

					err = fmt.Errorf("panic in UDF (func type %s): %v\nParameters: %v\nStack trace:\n%s",
						asf.userFunc.Type().String(), r, argValues, stackTrace)
				}
			}()

			// Total number of formal parameters in the function signature
			numFormalGoParams := len(asf.goArgTypes)
			numFixedGoParams := numFormalGoParams
			if asf.isVariadic {
				numFixedGoParams-- // Last formal parameter is the variadic slice itself
			}

			// Process arguments and call the function
			var callArgs []reflect.Value
			var argsErr error

			if asf.isVariadic {
				callArgs, argsErr = asf.processVariadicArgs(inputArgs, numFixedGoParams)
			} else {
				callArgs, argsErr = asf.processNonVariadicArgs(inputArgs)
			}

			if argsErr != nil {
				return nil, argsErr
			}

			// Call the user function
			results := asf.userFunc.Call(callArgs)
			userReturnVal := results[0].Interface()

			// Return value handling (including previous fix for map return types)
			if asf.goReturnType.Kind() == reflect.Map {
				if _, ok := userReturnVal.(duckdb.Map); !ok && userReturnVal != nil {
					rvMap := reflect.ValueOf(userReturnVal)
					if rvMap.Kind() == reflect.Map {
						newDuckDBMap := make(duckdb.Map, rvMap.Len())
						iter := rvMap.MapRange()
						for iter.Next() {
							newDuckDBMap[iter.Key().Interface()] = iter.Value().Interface()
						}
						return newDuckDBMap, nil
					}
				}
			}
			return userReturnVal, nil
		},
	}
}
