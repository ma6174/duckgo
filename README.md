# DuckGo: Create DuckDB UDFs in Go/XGo with Ease

[![Go Reference](https://pkg.go.dev/badge/github.com/ma6174/duckgo.svg)](https://pkg.go.dev/github.com/ma6174/duckgo)

`DuckGo` is a Go library designed to dramatically simplify the process of creating User-Defined Functions (UDFs) for [DuckDB](https://duckdb.org/). It uses Go's reflection mechanism to seamlessly convert a native Go function into a DuckDB scalar UDF. Additionally, it leverages [ixgo](https://github.com/goplus/ixgo) to support dynamically loading UDFs from Go/XGo scripts without prior compilation.

## Core Features

- **Create from Native Go Functions**: Directly convert your Go functions (e.g., `func(a, b int) int`) into DuckDB UDFs.
- **Automatic Type Mapping**: Automatically handles type conversions between Go and DuckDB, supporting a wide range of data types.
- **Variadic Function Support**: Seamlessly supports Go's variadic functions.
- **Dynamic Script Loading**: No compilation needed! Directly load functions from `.go` or `.xgo` source files as UDFs.
- **Direct Loading from SQL**: Provides a helper function to load and register UDFs from scripts directly within SQL queries.
- **Panic Handling**: Gracefully recovers from panics during UDF execution and converts them into DuckDB errors.

## Installation

```bash
go get github.com/ma6174/duckgo
```

## Important Note: Regarding Dynamic Scripting

The dynamic script loading feature (provided by the `script` package) relies on [ixgo](https://github.com/goplus/ixgo). Due to how `ixgo` works, you **must** add a specific linker flag to disable symbol name checking whenever you build or run code that uses the `script` package.

Failure to do so will result in a linker error.

**Build:**
```bash
go build -ldflags="-checklinkname=0" .
```

**Run:**
```bash
go run -ldflags="-checklinkname=0" .
```

If you import the `script` package in your own project, make sure to include this flag in your build and run commands.

## Usage

Here are a few examples of how to use `DuckGo`. For complete, runnable code, please see the [`example`](./example) directory.

### Example 1: Creating a UDF from a Native Go Function

This is the most basic use case. Any regular Go function can be registered.

The core logic involves wrapping the native Go function with `udf.BuildScalarUDF` and then registering it using `duckdb.RegisterScalarUDF`.

**For the full code, see: [`example/simple_udf/main.go`](./example/simple_udf/main.go)**

### Example 2: Loading a UDF from a Go Script

This is one of `DuckGo`'s most powerful features. You can register a function from a Go source file as a UDF without compiling it first.

The `script.AddIXGoUDFFromFile` function allows you to specify a `.go` file and the names of the functions you want to load.

**For the full code, see: [`example/script_udf/`](./example/script_udf/)**

### Example 3: Loading a UDF Directly via SQL

For even greater flexibility, you can call a helper function from within SQL to load UDFs.

First, enable the feature by calling `script.EnableRegisterUDFFromSQL(db)`. You can then use the `add_ixgo_udf` function in your SQL queries.

**For the full code, see: [`example/sql_load_udf/`](./example/sql_load_udf/)**

### Example 4: Importing an External Package

`DuckGo` supports importing functions from external packages (e.g., `github.com/goccy/go-json`) as UDFs. This is made possible by using `go generate` with [qexp](https://github.com/goplus/ixgo/tree/main/cmd/qexp) to pre-process the external package.

**For the full code, see: [`example/external_package/`](./example/external_package/)**

## Package Overview

- **`udf`**: The core package, responsible for converting native Go functions into DuckDB UDFs.
- **`script`**: Provides the functionality for dynamically loading UDFs from Go/XGo scripts.

## License

This project is licensed under the [MIT](LICENSE) License. 