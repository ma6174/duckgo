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

	// Enable the SQL function 'add_ixgo_udf'
	err = script.EnableRegisterUDFFromSQL(db)
	if err != nil {
		log.Fatal(err)
	}

	// Use SQL to load UDFs from a script
	// add_ixgo_udf(filename, function_name1, function_name2, ...)
	_, err = db.Exec("SELECT add_ixgo_udf('string_util.xgo', 'concatenate')")
	if err != nil {
		log.Fatal(err)
	}

	// Now you can use it!
	var result string
	err = db.QueryRow("SELECT concatenate('hello', 'world', '---')").Scan(&result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("concatenate('hello', 'world', '---') = %s\n", result)
}
