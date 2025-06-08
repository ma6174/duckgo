package udf

import (
	"context"
	"database/sql"
	"reflect"
	"strings"
	"testing"
)

// Helper functions only used in tests to execute queries and check results

// querySingleValueOnConn is a helper function to query a single value on a specific *sql.Conn
func querySingleValueOnConn(t *testing.T, conn *sql.Conn, query string, args ...interface{}) interface{} {
	t.Helper()
	var resultPlaceholder interface{} // Placeholder for Scan
	err := conn.QueryRowContext(context.Background(), query, args...).Scan(&resultPlaceholder)
	if err != nil {
		t.Fatalf("Error executing query: %v", err)
	}
	return resultPlaceholder
}

// If expected value is nil, we scan into *string to correctly handle SQL NULL vs. empty string

// For all integer types, scan into int64 as that's the generic representation from the DB

// expectQueryErrorOnConn is a helper function to expect an error on a specific *sql.Conn
func expectQueryErrorOnConn(t *testing.T, conn *sql.Conn, expectedErrorSubstring string, query string, args ...interface{}) {
	t.Helper()
	rows, errQuery := conn.QueryContext(context.Background(), query, args...)

	if errQuery == nil { // If query itself succeeds, try scanning to trigger an error from UDF
		defer rows.Close()
		if rows.Next() {
			var i int
			errScan := rows.Scan(&i)
			if errScan == nil {
				t.Errorf("Query '%s' expected to fail, error containing '%s', but succeeded and scanned successfully", query, expectedErrorSubstring)
				return
			}
			if !strings.Contains(errScan.Error(), expectedErrorSubstring) {
				t.Errorf("Query '%s' scanned error '%v' does not contain expected substring '%s'", query, errScan, expectedErrorSubstring)
			}
		} else if rows.Err() != nil {
			// Error during rows.Next() or rows.Err()
			if !strings.Contains(rows.Err().Error(), expectedErrorSubstring) {
				t.Errorf("Query '%s' rows.Err() '%v' does not contain expected substring '%s'", query, rows.Err(), expectedErrorSubstring)
			}
		} else {
			t.Errorf("Query '%s' succeeded but no rows returned, expected error containing '%s'", query, expectedErrorSubstring)
		}
		return
	}
	// Error from QueryContext itself
	if !strings.Contains(errQuery.Error(), expectedErrorSubstring) {
		t.Errorf("Query '%s' QueryContext error '%v' does not contain expected substring '%s'", query, errQuery, expectedErrorSubstring)
	}
}

// expectError is a helper function to verify error messages
func expectError(t *testing.T, err error, expectedMsgSubstring string) {
	t.Helper()
	if err == nil {
		t.Errorf("Expected error containing '%s', but got nil", expectedMsgSubstring)
		return
	}
	if !strings.Contains(err.Error(), expectedMsgSubstring) {
		t.Errorf("Error message '%s' does not contain expected substring '%s'", err.Error(), expectedMsgSubstring)
	}
}

// Helper function to check if a value is nil
func assertNil(t *testing.T, v interface{}, msg string, args ...interface{}) {
	t.Helper()
	if v != nil {
		t.Fatalf(msg, args...)
	}
}

// Helper function to check if two values are equal
func assertEqual(t *testing.T, expected, actual interface{}, msg string, args ...interface{}) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf(msg, args...)
	}
}

// Helper function to check if a condition is true
func assertTrue(t *testing.T, condition bool, msg string, args ...interface{}) {
	t.Helper()
	if !condition {
		t.Fatalf(msg, args...)
	}
}

// Helper function to fail the test with a message
func fail(t *testing.T, msg string, args ...interface{}) {
	t.Helper()
	t.Fatalf(msg, args...)
}
