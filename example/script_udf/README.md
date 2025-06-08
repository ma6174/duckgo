# Example: Dynamic UDF from Script

This example demonstrates how to dynamically load a User-Defined Function (UDF) into DuckDB from a Go source file (`my_udfs.go`) without prior compilation.

## How to Run

Due to the use of `ixgo` for dynamic code interpretation, you need to provide a specific linker flag to disable symbol name checking when running this example.

Use the following command from within this directory (`example/script_udf`):

```bash
go run -ldflags="-checklinkname=0" .
```

This will execute the `main.go` program, which dynamically loads the `my_multiply` function from `my_udfs.go` and then calls it from within a SQL query.
