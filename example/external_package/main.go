package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/ma6174/duckgo/example/external_package/github.com/goccy/go-json"
	"github.com/ma6174/duckgo/script" // Import the script package
	_ "github.com/marcboeker/go-duckdb/v2"
)

//go:generate go run github.com/goplus/ixgo/cmd/qexp -outdir . github.com/goccy/go-json

func main() {
	db, err := sql.Open("duckdb", "")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = script.AddIXGoUDFFromFile(db, "json.xgo", "jsonadd")
	if err != nil {
		log.Fatal(err)
	}

	var result int
	err = db.QueryRow("SELECT jsonadd('[1,2,3]')").Scan(&result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("jsonadd('[1,2,3]') = %v\n", result)
}
