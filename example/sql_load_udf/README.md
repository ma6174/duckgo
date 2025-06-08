# Example: Loading UDF via SQL

This example demonstrates how to enable a helper function in SQL (`add_ixgo_udf`) to dynamically load UDFs from Go source files.

## How to Run

This example also relies on `ixgo` for its dynamic capabilities. Therefore, you must use the `-ldflags="-checklinkname=0"` linker flag when running the code to avoid issues with symbol name checking.

Use the following command from within this directory (`example/sql_load_udf`):

```bash
go run -ldflags="-checklinkname=0" .
```

This command will execute `main.go`, which first registers the `add_ixgo_udf` helper function. It then uses this SQL function to load `concatenate` from `string_util.go` and executes it. 