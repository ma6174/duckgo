# Example: Dynamic UDF from Script

This example demonstrates how to dynamically load a User-Defined Function (UDF) into DuckDB from a Go source file (`my_udfs.go`) without prior compilation.

## How to Run

Use the following command from within this directory (`example/script_udf`):

```bash
go run .
```

This will execute the `main.go` program, which dynamically loads the `my_multiply` function from `my_udfs.go` and then calls it from within a SQL query.
