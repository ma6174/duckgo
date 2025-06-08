package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/ma6174/duckgo/script"
	_ "github.com/marcboeker/go-duckdb/v2"
)

func main() {
	db, err := sql.Open("duckdb", "")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 启用 SQL 函数 'add_ixgo_udf'
	err = script.EnableRegisterUDFFromSQL(db)
	if err != nil {
		log.Fatal(err)
	}

	// 使用 SQL 来加载脚本中的 UDF
	// add_ixgo_udf(文件名, 函数名1, 函数名2, ...)
	_, err = db.Exec("SELECT add_ixgo_udf('string_util.xgo', 'concatenate')")
	if err != nil {
		log.Fatal(err)
	}

	// 现在就可以使用了!
	var result string
	err = db.QueryRow("SELECT concatenate('hello', 'world', '---')").Scan(&result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("concatenate('hello', 'world', '---') = %s\n", result)
}
