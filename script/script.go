package script

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/goplus/ixgo"
	_ "github.com/goplus/ixgo/pkg"
	_ "github.com/goplus/ixgo/xgobuild"
	"github.com/ma6174/duckgo/udf"
	"github.com/marcboeker/go-duckdb/v2"
)

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

func AddIXGoUDFFromFile(db *sql.DB, filename string, funcNames ...string) (err error) {
	return addIXGoUDF(db, filename, nil, funcNames...)
}

func AddIXGoUDFFromSource(db *sql.DB, src any, funcNames ...string) (err error) {
	return addIXGoUDF(db, "main.xgo", src, funcNames...)
}

func addIXGoUDF(db *sql.DB, filename string, src any, funcNames ...string) (err error) {
	conn, err := db.Conn(context.Background())
	if err != nil {
		return err
	}
	defer conn.Close()
	ctx := ixgo.NewContext(0)
	pkg, err := ctx.LoadFile(filename, src)
	if err != nil {
		return
	}
	interp, err := ctx.NewInterp(pkg)
	if err != nil {
		return
	}
	err = interp.RunInit()
	if err != nil {
		return
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
