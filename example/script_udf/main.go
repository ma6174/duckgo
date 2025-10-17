package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/duckdb/duckdb-go/v2"
	"github.com/ma6174/duckgo/script" // Import the script package
)

func main() {
	db, err := sql.Open("duckdb", "")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Load the "multiply" function from the script file
	err = script.AddIXGoUDFFromFile(db, "my_udfs.go", "my_multiply")
	if err != nil {
		log.Fatal(err)
	}

	// Use it in SQL
	var result int
	err = db.QueryRow("SELECT my_multiply(7, 6)").Scan(&result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("my_multiply(7, 6) = %d\n", result)
}
