package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/ma6174/duckgo/udf"
	"github.com/marcboeker/go-duckdb/v2"
)

func main() {
	db, err := sql.Open("duckdb", "")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 1. 定义一个你想要在 SQL 中使用的 Go 函数
	add := func(a, b int) int {
		return a + b
	}

	// 2. 使用 udf.BuildScalarUDF 将其包装成 DuckDB UDF
	scalarFunc, err := udf.BuildScalarUDF(add)
	if err != nil {
		log.Fatal(err)
	}

	// 3. 获取连接并注册 UDF
	conn, err := db.Conn(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	err = duckdb.RegisterScalarUDF(conn, "go_add", scalarFunc)
	if err != nil {
		log.Fatal(err)
	}

	// 4. 现在你可以在 SQL 中调用它了!
	var result int
	err = db.QueryRow("SELECT go_add(5, 3)").Scan(&result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("go_add(5, 3) = %d\n", result)
}
