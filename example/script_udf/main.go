package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/ma6174/duckgo/script" // 导入 script 包
	_ "github.com/marcboeker/go-duckdb/v2"
)

func main() {
	db, err := sql.Open("duckdb", "")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 从脚本文件加载 "multiply" 函数
	err = script.AddIXGoUDFFromFile(db, "my_udfs.go", "my_multiply")
	if err != nil {
		log.Fatal(err)
	}

	// 在 SQL 中使用它
	var result int
	err = db.QueryRow("SELECT my_multiply(7, 6)").Scan(&result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("my_multiply(7, 6) = %d\n", result)
}
