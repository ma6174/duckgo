package udf

import (
	"testing"
)

func TestBuildScalarUDFErrors(t *testing.T) {
	tests := []struct {
		name               string
		udfSQLName         string
		fn                 any
		options            []func(*udfOption)
		expectedErrMessage string
	}{
		{
			name:               "non-function input",
			udfSQLName:         "test_non_func",
			fn:                 123,
			options:            nil,
			expectedErrMessage: "is not a function",
		},
		{
			name:               "no return value",
			udfSQLName:         "test_no_ret",
			fn:                 func() {},
			options:            nil,
			expectedErrMessage: "must return exactly one value",
		},
		{
			name:               "multiple return values",
			udfSQLName:         "test_multi_ret",
			fn:                 func() (int, int) { return 1, 2 },
			options:            nil,
			expectedErrMessage: "must return exactly one value",
		},
		{
			name:               "unsupported arg type",
			udfSQLName:         "test_unsupp_arg",
			fn:                 func(a chan int) int { return 1 },
			options:            nil,
			expectedErrMessage: "unsupported Go type kind for UDF: chan (specific type: chan int)",
		},
		{
			name:               "unsupported return type",
			udfSQLName:         "test_unsupp_ret",
			fn:                 func() chan int { return make(chan int) },
			options:            nil,
			expectedErrMessage: "unsupported Go type kind for UDF: chan (specific type: chan int)",
		},
		{
			name:               "unsupported variadic element type",
			udfSQLName:         "test_unsupp_var_elem",
			fn:                 func(fixed string, variadic ...chan int) string { return fixed },
			options:            nil,
			expectedErrMessage: "BuildScalarUDF: error converting Go variadic element type for UDF (Go type chan int, func type func(string, ...chan int) string): unsupported Go type kind for UDF: chan (specific type: chan int)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := BuildScalarUDF(tt.fn, tt.options...)
			expectError(t, err, tt.expectedErrMessage)
		})
	}
}
