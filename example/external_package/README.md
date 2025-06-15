# Example: Import External Package

This example demonstrates how to import an external Go package (`github.com/goccy/go-json`) and use its functions as User-Defined Functions (UDFs) in DuckDB.

## How to Run

1.  **Generate Code**

    First, you need to run the `go generate` command. This command uses [qexp](https://github.com/goplus/ixgo/tree/main/cmd/qexp) to process the external package, enabling it to be interpreted dynamically.

    ```bash
    go generate
    ```

2.  **Run the Example**

    Due to the use of `ixgo` for dynamic code interpretation, you need to provide a specific linker flag (`-checklinkname=0`) to disable symbol link name checking when running the example.

    Execute the following command from within this directory (`example/external_package`):

    ```bash
    go run -ldflags="-checklinkname=0" .
    ```

    This will execute the `main.go` program, which dynamically loads functions from the `github.com/goccy/go-json` package and calls them within a SQL query.
