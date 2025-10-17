package script

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/duckdb/duckdb-go/v2"
	"github.com/goplus/ixgo"
	_ "github.com/goplus/ixgo/pkg"
	_ "github.com/goplus/ixgo/xgobuild"
	"github.com/ma6174/duckgo/udf"
)

// EnableRegisterUDFFromSQL enables the registration of user-defined functions (UDFs) from SQL.
// It registers a scalar UDF named "add_ixgo_udf" that allows adding more UDFs
// from .ixgo files directly within SQL queries.
// The signature of the SQL function is add_ixgo_udf(filename TEXT, funcNames TEXT...).
func EnableRegisterUDFFromSQL(db *sql.DB) (err error) {
	addUDF := func(filename string, funcNames ...string) int {
		err = AddIXGoUDFFromFile(db, filename, funcNames...)
		if err != nil {
			log.Panicln(err)
		}
		return 1
	}
	sf, err := udf.BuildScalarUDF(addUDF)
	if err != nil {
		return err
	}
	conn, err := db.Conn(context.Background())
	if err != nil {
		return err
	}
	defer conn.Close()
	return duckdb.RegisterScalarUDF(conn, "add_ixgo_udf", sf)
}

// AddIXGoUDFFromFile loads an .go or .xgo script from a file and registers the specified functions as UDFs in DuckDB.
func AddIXGoUDFFromFile(db *sql.DB, filename string, funcNames ...string) (err error) {
	return addIXGoUDF(db, filename, nil, funcNames...)
}

// AddIXGoUDFFromSource loads an .go or .xgo script from a source string or byte slice and registers the specified functions as UDFs in DuckDB.
// The filename is used for error reporting.
func AddIXGoUDFFromSource(db *sql.DB, src any, funcNames ...string) (err error) {
	return addIXGoUDF(db, "main.xgo", src, funcNames...)
}

// addIXGoUDF is an internal function that handles the logic for loading an .go or .xgo package
// from either a file or source, interpreting it, and registering the specified functions
// as scalar UDFs in DuckDB.
func addIXGoUDF(db *sql.DB, filename string, src any, funcNames ...string) (err error) {
	conn, err := db.Conn(context.Background())
	if err != nil {
		return err
	}
	defer conn.Close()
	ctx := ixgo.NewContext(0)
	pkg, err := ctx.LoadFile(filename, src)
	if err != nil {
		return err
	}
	interp, err := ctx.NewInterp(pkg)
	if err != nil {
		return err
	}
	err = interp.RunInit()
	if err != nil {
		return err
	}
	for _, funcName := range funcNames {
		fi, ok := interp.GetFunc(funcName)
		if !ok {
			return errors.New("func not found")
		}
		sf, err := udf.BuildScalarUDF(fi)
		if err != nil {
			return err
		}
		log.Println("AddIXGoUDF", filename, funcName)
		err = duckdb.RegisterScalarUDF(conn, funcName, sf)
		if err != nil {
			return err
		}
	}
	return nil
}
