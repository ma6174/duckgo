package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/duckdb/duckdb-go/v2"
	"github.com/ma6174/duckgo/udf"
)

func main() {
	db, err := sql.Open("duckdb", "")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 1. Define a Go function that you want to use in SQL
	add := func(a, b int) int {
		return a + b
	}

	// 2. Wrap it as a DuckDB UDF using udf.BuildScalarUDF
	scalarFunc, err := udf.BuildScalarUDF(add)
	if err != nil {
		log.Fatal(err)
	}

	// 3. Get a connection and register the UDF
	conn, err := db.Conn(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	err = duckdb.RegisterScalarUDF(conn, "go_add", scalarFunc)
	if err != nil {
		log.Fatal(err)
	}

	// 4. Now you can call it in SQL!
	var result int
	err = db.QueryRow("SELECT go_add(5, 3)").Scan(&result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("go_add(5, 3) = %d\n", result)
}
