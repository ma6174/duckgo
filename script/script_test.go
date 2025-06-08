package script

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/ma6174/duckgo/udf"
	"github.com/marcboeker/go-duckdb/v2"
	"github.com/stretchr/testify/require"
)

func newTestDB(t testing.TB) *sql.DB {
	db, err := sql.Open("duckdb", "")
	require.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})
	return db
}

func TestAddIXGoUDF(t *testing.T) {
	tempDir := t.TempDir()
	udfFile := filepath.Join(tempDir, "udf.go")
	err := os.WriteFile(udfFile, []byte(`
package main

func myadd(a,b int) int {
	return a+b
}
`), 0644)
	require.NoError(t, err)

	t.Run("from file", func(t *testing.T) {
		db := newTestDB(t)
		err := AddIXGoUDFFromFile(db, udfFile, "myadd")
		require.NoError(t, err)
		var result int
		err = db.QueryRow("select myadd(1,2)").Scan(&result)
		require.NoError(t, err)
		require.Equal(t, 3, result)
	})

	t.Run("from source", func(t *testing.T) {
		db := newTestDB(t)
		src := `
package main

func mysub(a,b int) int {
	return a-b
}`
		err := AddIXGoUDFFromSource(db, src, "mysub")
		require.NoError(t, err)
		var result int
		err = db.QueryRow("select mysub(3,1)").Scan(&result)
		require.NoError(t, err)
		require.Equal(t, 2, result)
	})
}

func TestEnableRegisterUDFFromSQL(t *testing.T) {
	tempDir := t.TempDir()
	udfFile := filepath.Join(tempDir, "udf.go")
	err := os.WriteFile(udfFile, []byte(`
package main

func mymul(a,b int) int {
	return a*b
}
`), 0644)
	require.NoError(t, err)

	db := newTestDB(t)
	err = EnableRegisterUDFFromSQL(db)
	require.NoError(t, err)

	_, err = db.Exec("select add_ixgo_udf(?, 'mymul')", udfFile)
	require.NoError(t, err)

	var result int
	err = db.QueryRow("select mymul(3,4)").Scan(&result)
	require.NoError(t, err)
	require.Equal(t, 12, result)
}

func TestAddUDF(t *testing.T) {
	c, err := duckdb.NewConnector("", nil)
	require.NoError(t, err)

	db := sql.OpenDB(c)
	defer db.Close()
	EnableRegisterUDFFromSQL(db)
	sf, err := udf.BuildScalarUDF(func(a int, b string) string {
		return fmt.Sprintf("%v%v", a, b)
	})
	require.NoError(t, err)
	conn, err := db.Conn(context.Background())
	require.NoError(t, err)
	defer conn.Close()
	err = duckdb.RegisterScalarUDF(conn, "go_test_udf", sf)
	require.NoError(t, err)
}

func BenchmarkNativeUDF(b *testing.B) {
	db := newTestDB(b)
	add := func(a, b int) int {
		return a + b
	}
	sf, err := udf.BuildScalarUDF(add)
	require.NoError(b, err)
	conn, err := db.Conn(context.Background())
	require.NoError(b, err)
	defer conn.Close()
	err = duckdb.RegisterScalarUDF(conn, "go_add", sf)
	require.NoError(b, err)

	b.ResetTimer()
	var result int
	for i := 0; i < b.N; i++ {
		err := db.QueryRow("SELECT go_add(1, 2)").Scan(&result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIXGoUDF(b *testing.B) {
	db := newTestDB(b)
	src := `
package main

func ixgo_add(a,b int) int {
	return a+b
}`
	err := AddIXGoUDFFromSource(db, src, "ixgo_add")
	require.NoError(b, err)

	b.ResetTimer()
	var result int
	for i := 0; i < b.N; i++ {
		err := db.QueryRow("SELECT ixgo_add(1, 2)").Scan(&result)
		if err != nil {
			b.Fatal(err)
		}
	}
}
