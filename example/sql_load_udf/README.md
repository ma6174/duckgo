# Example: Loading UDF via SQL

This example demonstrates how to enable a helper function in SQL (`add_ixgo_udf`) to dynamically load UDFs from Go source files.

## How to Run

Use the following command from within this directory (`example/sql_load_udf`):

```bash
go run .
```

This command will execute `main.go`, which first registers the `add_ixgo_udf` helper function. It then uses this SQL function to load `concatenate` from `string_util.go` and executes it. 